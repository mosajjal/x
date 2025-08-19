package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	PcapMagicLE  = 0xa1b2c3d4
	PcapMagicBE  = 0xd4c3b2a1
	MinTimestamp = 631152000  // ~1990
	MaxTimestamp = 2208988800 // ~2040
	MaxPacketLen = 65536
)

type PcapResult struct {
	FirstTimestamp time.Time
	LastTimestamp  time.Time
	FirstTime      time.Time
	LastTime       time.Time
	Duration       time.Duration
}

func readPcapTimestamps(filename string) (*PcapResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read global header (24 bytes)
	globalHeader := make([]byte, 24)
	if _, err := io.ReadFull(file, globalHeader); err != nil {
		return nil, fmt.Errorf("failed to read global header: %v", err)
	}

	// Determine endianness from magic number
	magic := binary.LittleEndian.Uint32(globalHeader[:4])
	var byteOrder binary.ByteOrder

	switch magic {
	case PcapMagicLE:
		byteOrder = binary.LittleEndian
	case PcapMagicBE:
		byteOrder = binary.BigEndian
	default:
		return nil, fmt.Errorf("invalid pcap magic number: 0x%x", magic)
	}

	// Read first packet header (16 bytes)
	packetHeader := make([]byte, 16)
	if _, err := io.ReadFull(file, packetHeader); err != nil {
		return nil, fmt.Errorf("no packets in file or corrupt header: %v", err)
	}

	// Extract first packet timestamp
	tsSec := byteOrder.Uint32(packetHeader[0:4])
	tsUsec := byteOrder.Uint32(packetHeader[4:8])
	firstTimestamp := float64(tsSec) + float64(tsUsec)/1000000.0

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := fileInfo.Size()

	// Find last packet by reading from the tail
	lastTimestamp := firstTimestamp

	// Read last 64KB or entire file if smaller
	tailSize := int64(65536)
	if tailSize > fileSize-24 {
		tailSize = fileSize - 24
	}

	tailStart := fileSize - tailSize
	if _, err := file.Seek(tailStart, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to tail: %v", err)
	}

	tailData := make([]byte, tailSize)
	if _, err := io.ReadFull(file, tailData); err != nil {
		return nil, fmt.Errorf("failed to read tail data: %v", err)
	}

	// Parse through tail data looking for packet headers
	lastTimestamp = parsePacketsInBuffer(tailData, byteOrder, lastTimestamp)

	// If we didn't find packets in tail, scan the entire file
	if lastTimestamp == firstTimestamp && fileSize > 24+tailSize {
		if _, err := file.Seek(24, 0); err != nil {
			return nil, fmt.Errorf("failed to seek to start: %v", err)
		}
		lastTimestamp = scanFileForLastTimestamp(file, byteOrder, fileSize, lastTimestamp)
	}

	// Convert to time.Time
	firstTime := time.Unix(int64(firstTimestamp), int64((firstTimestamp-float64(int64(firstTimestamp)))*1e9))
	lastTime := time.Unix(int64(lastTimestamp), int64((lastTimestamp-float64(int64(lastTimestamp)))*1e9))

	return &PcapResult{
		FirstTimestamp: firstTime,
		LastTimestamp:  lastTime,
		FirstTime:      firstTime,
		LastTime:       lastTime,
		Duration:       lastTime.Sub(firstTime),
	}, nil
}

func parsePacketsInBuffer(data []byte, byteOrder binary.ByteOrder, currentLastTimestamp float64) float64 {
	lastTimestamp := currentLastTimestamp
	i := 0

	for i < len(data)-16 {
		if i+16 > len(data) {
			break
		}

		// Read potential packet header
		tsSec := byteOrder.Uint32(data[i : i+4])
		tsUsec := byteOrder.Uint32(data[i+4 : i+8])
		capLen := byteOrder.Uint32(data[i+8 : i+12])
		origLen := byteOrder.Uint32(data[i+12 : i+16])

		// Validate packet header
		if isValidPacketHeader(tsSec, tsUsec, capLen, origLen) &&
			i+16+int(capLen) <= len(data) {

			timestamp := float64(tsSec) + float64(tsUsec)/1000000.0
			if timestamp > lastTimestamp {
				lastTimestamp = timestamp
			}

			// Move to next packet
			i += 16 + int(capLen)
		} else {
			i++
		}
	}

	return lastTimestamp
}

func scanFileForLastTimestamp(file *os.File, byteOrder binary.ByteOrder, fileSize int64, currentLastTimestamp float64) float64 {
	lastTimestamp := currentLastTimestamp
	chunkSize := int64(1024 * 1024) // 1MB chunks
	buffer := make([]byte, chunkSize)

	for {
		pos, err := file.Seek(0, 1) // Get current position
		if err != nil || pos >= fileSize-16 {
			break
		}

		readSize := chunkSize
		if pos+readSize > fileSize {
			readSize = fileSize - pos
		}

		n, err := file.Read(buffer[:readSize])
		if err != nil || n < 16 {
			break
		}

		// Process chunk
		i := 0
		for i < n-16 {
			tsSec := byteOrder.Uint32(buffer[i : i+4])
			tsUsec := byteOrder.Uint32(buffer[i+4 : i+8])
			capLen := byteOrder.Uint32(buffer[i+8 : i+12])
			origLen := byteOrder.Uint32(buffer[i+12 : i+16])

			if isValidPacketHeader(tsSec, tsUsec, capLen, origLen) {
				timestamp := float64(tsSec) + float64(tsUsec)/1000000.0
				if timestamp > lastTimestamp {
					lastTimestamp = timestamp
				}

				// Skip to next packet
				packetEnd := i + 16 + int(capLen)
				if packetEnd >= n {
					// Packet extends beyond chunk, seek to continue
					if _, err := file.Seek(pos+int64(packetEnd), 0); err != nil {
						return lastTimestamp
					}
					break
				}
				i = packetEnd
			} else {
				i++
			}
		}
	}

	return lastTimestamp
}

func isValidPacketHeader(tsSec, tsUsec, capLen, origLen uint32) bool {
	return tsSec >= MinTimestamp &&
		tsSec <= MaxTimestamp &&
		tsUsec < 1000000 &&
		capLen > 0 &&
		capLen <= MaxPacketLen &&
		capLen <= origLen
}
