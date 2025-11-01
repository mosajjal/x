# AllHash

[![Go Reference](https://pkg.go.dev/badge/github.com/mosajjal/go-exp/allhash.svg)](https://pkg.go.dev/github.com/mosajjal/go-exp/allhash)

Multi-hash calculator for forensic analysis and file verification. Generates multiple cryptographic and fuzzy hashes in a single pass for efficient file analysis.

## Features

- 🔐 **Cryptographic Hashes**: MD5, SHA1, SHA256, SHA512
- 🔍 **Fuzzy Hashes**: TLSH (Trend Micro Locality Sensitive Hash)
- 📊 **Similarity Detection**: ssdeep (context-triggered piecewise hashing)
- ⚡ **Single Pass**: Calculates all hashes in one file read
- 💾 **Memory Efficient**: Streaming hash calculation

## Installation

```bash
go install github.com/mosajjal/go-exp/allhash@latest
```

Or build from source:

```bash
git clone https://github.com/mosajjal/go-exp.git
cd go-exp/allhash
go build
```

## Usage

### Basic Usage

Calculate all hashes for a file:

```bash
allhash /path/to/file
```

### Example Output

```
File: malware.exe
MD5:      5d41402abc4b2a76b9719d911017c592
SHA1:     aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
SHA256:   2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae
SHA512:   cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce...
TLSH:     T1A5B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2
ssdeep:   3:AE:A
```

## Hash Types

### MD5 (Message Digest 5)
- **Size**: 128-bit (32 hex characters)
- **Use**: Legacy file identification, checksums
- **Security**: Cryptographically broken, not for security purposes

### SHA1 (Secure Hash Algorithm 1)
- **Size**: 160-bit (40 hex characters)
- **Use**: Legacy file verification, Git commits
- **Security**: Deprecated for cryptographic use

### SHA256 (SHA-2)
- **Size**: 256-bit (64 hex characters)
- **Use**: Modern file verification, digital signatures
- **Security**: Currently secure

### SHA512 (SHA-2)
- **Size**: 512-bit (128 hex characters)
- **Use**: High-security file verification
- **Security**: Currently secure, higher collision resistance

### TLSH (Trend Micro Locality Sensitive Hash)
- **Purpose**: Fuzzy matching, similarity detection
- **Use**: Malware variant detection, near-duplicate finding
- **Feature**: Can measure similarity between files

### ssdeep (Context-Triggered Piecewise Hashing)
- **Purpose**: Piecewise hash for similarity detection
- **Use**: Finding similar files, malware analysis
- **Feature**: Detects files with shared sequences

## Use Cases

### Digital Forensics

```bash
# Hash evidence files
for file in evidence/*; do
  allhash "$file" >> evidence_hashes.txt
done
```

### Malware Analysis

```bash
# Compare sample with known malware database
allhash suspicious.exe | grep -f known_malware_hashes.txt
```

### File Integrity

```bash
# Create baseline
allhash /etc/passwd > passwd.hashes

# Later, verify
allhash /etc/passwd | diff - passwd.hashes
```

### Duplicate Detection

```bash
# Find similar files using ssdeep
allhash file1.dat | grep ssdeep > hash1.txt
allhash file2.dat | grep ssdeep > hash2.txt
ssdeep -l -r hash1.txt hash2.txt
```

## Comparison with VirusTotal

AllHash generates the same hash types provided by VirusTotal for uploaded files:
- MD5, SHA1, SHA256 for exact matching
- TLSH, ssdeep for fuzzy matching and variant detection

## Performance

- **Single File**: < 1 second for typical executables
- **Large Files**: Processes at disk I/O speed
- **Memory Usage**: Constant (not dependent on file size)

## Output Format

Plain text with labeled hash values:

```
File: example.bin
MD5: <hash>
SHA1: <hash>
SHA256: <hash>
SHA512: <hash>
TLSH: <hash>
ssdeep: <hash>
```

## Integration Examples

### With jq

```bash
# Convert to JSON
allhash file.exe | awk -F': ' '{print "\"" $1 "\": \"" $2 "\","}' | \
  sed '1s/^/{\n/; $s/,$/\n}/'
```

### With CSV

```bash
# Create CSV of hashes
echo "filename,md5,sha256" > hashes.csv
for file in *; do
  echo -n "$file," >> hashes.csv
  allhash "$file" | grep -E "^(MD5|SHA256)" | cut -d: -f2 | \
    tr '\n' ',' | sed 's/,$/\n/' >> hashes.csv
done
```

### Batch Processing

```bash
# Hash all executables
find /usr/bin -type f -executable | \
  xargs -I {} sh -c 'echo "=== {} ===" && allhash {}'
```

## Requirements

- Go 1.24 or later
- Sufficient permissions to read target files

## Limitations

- No directory recursion (use with `find` or `xargs`)
- No parallel processing of multiple files
- No streaming from stdin

## Contributing

Contributions welcome! Please see the main repository [CONTRIBUTING.md](../CONTRIBUTING.md).

## License

See [LICENSE](LICENSE) file in the project root.

## Related Tools

- [hashdeep](https://github.com/jessek/hashdeep) - Comprehensive file hashing
- [VirusTotal](https://www.virustotal.com/) - Online malware scanner
- [ssdeep](https://ssdeep-project.github.io/ssdeep/) - Fuzzy hashing tool

---

**Note**: For legal forensic work, ensure proper chain of custody and documentation.