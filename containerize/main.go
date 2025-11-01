package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"debug/elf"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DockerImageManifest represents the structure of a Docker image manifest.
type DockerImageManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// DockerConfig represents the Docker image configuration.
type DockerConfig struct {
	Architecture string          `json:"architecture"`
	OS           string          `json:"os"`
	Created      string          `json:"created"`
	Config       ContainerConfig `json:"config"`
	RootFS       RootFS          `json:"rootfs"`
}

// ContainerConfig holds container configuration.
type ContainerConfig struct {
	Entrypoint []string `json:"Entrypoint"`
}

// RootFS describes the root filesystem.
type RootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

func main() {
	// 1. Validate input
	if len(os.Args) < 2 {
		log.Fatal("Usage: go-oci-builder <path-to-static-binary>")
	}
	binaryPath := os.Args[1]
	binaryName := filepath.Base(binaryPath)

	// Detect the architecture of the binary
	arch, err := detectArchitecture(binaryPath)
	if err != nil {
		log.Fatalf("Failed to detect architecture of binary: %v", err)
	}
	log.Printf("Detected architecture: %s", arch)

	// Create a temporary working directory for building the image components.
	workDir, err := os.MkdirTemp("", "docker-image-builder-*")
	if err != nil {
		log.Fatalf("Failed to create temp working directory: %v", err)
	}
	defer os.RemoveAll(workDir) // Clean up when we're done.
	log.Printf("Using working directory: %s", workDir)

	// --- Step 2: Create the Root Filesystem (rootfs) Layer ---
	log.Println("Creating rootfs layer...")
	rootfsPath := filepath.Join(workDir, "rootfs")
	appDir := filepath.Join(rootfsPath, "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		log.Fatalf("Failed to create app directory: %v", err)
	}

	// Copy the user's binary into the rootfs.
	destBinaryPath := filepath.Join(appDir, binaryName)
	if err := copyFile(binaryPath, destBinaryPath); err != nil {
		log.Fatalf("Failed to copy binary: %v", err)
	}
	if err := os.Chmod(destBinaryPath, 0755); err != nil {
		log.Fatalf("Failed to set binary permissions: %v", err)
	} // Ensure the binary is executable.

	// Create a gzipped tarball of the rootfs. This is our single layer.
	layerTarGzPath := filepath.Join(workDir, "layer.tar.gz")
	layerDigest, layerDiffID, layerSize := createLayerTarball(rootfsPath, layerTarGzPath)
	log.Printf("Layer created: %s (Size: %d bytes, Digest: %s, DiffID: %s)", layerTarGzPath, layerSize, layerDigest, layerDiffID)

	// --- Step 3: Create the Image Configuration JSON ---
	log.Println("Creating image config JSON...")
	imageConfig := DockerConfig{
		Architecture: arch,
		OS:           "linux",
		Created:      time.Now().UTC().Format(time.RFC3339),
	}
	imageConfig.Config.Entrypoint = []string{filepath.Join("/app", binaryName)}
	imageConfig.RootFS.Type = "layers"
	imageConfig.RootFS.DiffIDs = []string{layerDiffID}

	configBytes, err := json.Marshal(imageConfig)
	if err != nil {
		log.Fatalf("Failed to marshal image config: %v", err)
	}
	configDigest, configSize := calculateDigestAndSize(configBytes)
	configFileName := fmt.Sprintf("%s.json", strings.TrimPrefix(configDigest, "sha256:"))
	configPath := filepath.Join(workDir, configFileName)
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}
	log.Printf("Config created: %s (Size: %d bytes, Digest: %s)", configPath, configSize, configDigest)

	// --- Step 4: Create the Manifest JSON ---
	log.Println("Creating manifest JSON...")
	layerFileName := fmt.Sprintf("%s.tar.gz", strings.TrimPrefix(layerDigest, "sha256:"))
	dockerManifest := []DockerImageManifest{
		{
			Config:   configFileName,
			RepoTags: []string{"go-containerized-app:latest"},
			Layers:   []string{layerFileName},
		},
	}

	manifestBytes, err := json.Marshal(dockerManifest)
	if err != nil {
		log.Fatalf("Failed to marshal manifest: %v", err)
	}
	manifestPath := filepath.Join(workDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		log.Fatalf("Failed to write manifest file: %v", err)
	}
	log.Printf("Manifest created: %s", manifestPath)

	// --- Step 5: Assemble the Docker Image Layout ---
	log.Println("Assembling Docker image layout...")

	// Copy config and layer with their digest-based names
	layerPath := filepath.Join(workDir, layerFileName)
	if err := copyFile(layerTarGzPath, layerPath); err != nil {
		log.Fatalf("Failed to copy layer: %v", err)
	}

	// --- Step 6: Create the Final Output Tarball ---
	log.Println("Creating final image.tar...")
	if err := createDockerTarball(workDir, "image.tar", []string{configFileName, layerFileName, "manifest.json"}); err != nil {
		log.Fatalf("Failed to create final tarball: %v", err)
	}
	log.Println("Successfully created image.tar!")
}

// createLayerTarball tars and gzips a directory, returning the digest of the compressed tarball,
// the digest of the uncompressed tarball (diffID), and its size.
func createLayerTarball(sourceDir, targetTarGz string) (string, string, int64) {
	// Create the target file
	outFile, err := os.Create(targetTarGz)
	if err != nil {
		log.Fatalf("Failed to create layer tarball file: %v", err)
	}
	defer outFile.Close()

	// Create writers
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk the source directory and add files to the tar archive
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a proper header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Calculate the relative path correctly
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// For Docker, we need to use "./" prefix for the root directory entries
		if relPath == "." {
			header.Name = "./"
		} else {
			header.Name = relPath
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to create layer tarball: %v", err)
	}

	// Must close writers to flush buffers before calculating digests
	tarWriter.Close()
	gzipWriter.Close()
	outFile.Close()

	// Calculate digest of the compressed tar.gz
	fileBytes, err := os.ReadFile(targetTarGz)
	if err != nil {
		log.Fatalf("Failed to read layer tarball for hashing: %v", err)
	}
	compressedDigest, compressedSize := calculateDigestAndSize(fileBytes)

	// Calculate diffID (digest of the uncompressed tar)
	// This is a bit more involved as we need to decompress in memory to hash.
	gzFile, err := os.Open(targetTarGz)
	if err != nil {
		log.Fatalf("Failed to open layer tarball: %v", err)
	}
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	uncompressedBytes, err := io.ReadAll(gzReader)
	if err != nil {
		log.Fatalf("Failed to read uncompressed layer: %v", err)
	}
	uncompressedDigest, _ := calculateDigestAndSize(uncompressedBytes)

	return compressedDigest, uncompressedDigest, compressedSize
}

// createDockerTarball creates a tar archive suitable for Docker load
func createDockerTarball(sourceDir, targetTar string, filesToInclude []string) error {
	outFile, err := os.Create(targetTar)
	if err != nil {
		return fmt.Errorf("failed to create final tarball: %v", err)
	}
	defer outFile.Close()

	tarWriter := tar.NewWriter(outFile)
	defer tarWriter.Close()

	// Add each specified file to the tar
	for _, fileName := range filesToInclude {
		filePath := filepath.Join(sourceDir, fileName)

		// Get file info
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %v", filePath, err)
		}

		// Create header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create header for %s: %v", filePath, err)
		}
		header.Name = fileName // Use just the filename, not the full path

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %v", filePath, err)
		}

		// Copy file content if it's not a directory
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", filePath, err)
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return fmt.Errorf("failed to copy file %s: %v", filePath, err)
			}
		}
	}

	return nil
}

// calculateDigestAndSize computes the SHA256 digest and size of a byte slice.
func calculateDigestAndSize(data []byte) (string, int64) {
	hash := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(hash[:])
	return digest, int64(len(data))
}

// copyFile is a simple utility to copy a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file from %s to %s: %v", src, dst, err)
	}

	return nil
}

// detectArchitecture determines the architecture of a binary file
func detectArchitecture(binaryPath string) (string, error) {
	// Try to open as ELF
	f, err := elf.Open(binaryPath)
	if err != nil {
		return "", fmt.Errorf("not a valid ELF binary: %v", err)
	}
	defer f.Close()

	// Map ELF machine type to Docker architecture
	switch f.Machine {
	case elf.EM_X86_64:
		return "amd64", nil
	case elf.EM_AARCH64:
		return "arm64", nil
	case elf.EM_ARM:
		return "arm", nil
	case elf.EM_386:
		return "386", nil
	case elf.EM_PPC64:
		if f.ByteOrder == binary.LittleEndian {
			return "ppc64le", nil
		}
		return "ppc64", nil
	case elf.EM_S390:
		return "s390x", nil
	case elf.EM_RISCV:
		switch f.Class {
		case elf.ELFCLASS32:
			return "riscv32", nil
		case elf.ELFCLASS64:
			return "riscv64", nil
		}
	}

	return "", fmt.Errorf("unsupported architecture: %s", f.Machine.String())
}
