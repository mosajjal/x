package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcap"
	"github.com/gopacket/gopacket/pcapgo"
)

func main() {
	// Define command-line flags
	inputFile := flag.String("i", "", "Input PCAP/PCAPNG file")
	outputFile := flag.String("o", "", "Output PCAP/PCAPNG file")
	timeShiftStr := flag.String("t", "now", "Time shift duration (e.g., '24h', '1h30m'). 'now' sets the latest packet to the current time.")
	flag.Parse()

	// Validate input and output file paths
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Input and output files must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	// Open the input file
	handle, err := pcap.OpenOffline(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer handle.Close()

	// Find the latest timestamp in the pcap file
	var latestTimestamp time.Time
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if packet.Metadata().Timestamp.After(latestTimestamp) {
			latestTimestamp = packet.Metadata().Timestamp
		}
	}

	// Calculate the time shift
	var timeShift time.Duration
	if *timeShiftStr == "now" {
		timeShift = time.Now().Sub(latestTimestamp)
	} else {
		timeShift, err = time.ParseDuration(*timeShiftStr)
		if err != nil {
			fmt.Printf("Error parsing time shift duration: %v\n", err)
			os.Exit(1)
		}
	}

	// Create the output file
	f, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Create a new pcap writer
	w := pcapgo.NewWriter(f)
	if err := w.WriteFileHeader(uint32(handle.SnapLen()), handle.LinkType()); err != nil {
		fmt.Printf("Error writing file header: %v\n", err)
		os.Exit(1)
	}

	// Reset the reader to the beginning of the file
	handle, err = pcap.OpenOffline(*inputFile)
	if err != nil {
		fmt.Printf("Error reopening input file: %v\n", err)
		os.Exit(1)
	}
	defer handle.Close()

	// Iterate through the packets again, adjust the timestamp, and write to the new file
	packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		newTime := packet.Metadata().Timestamp.Add(timeShift)
		ci := packet.Metadata().CaptureInfo
		ci.Timestamp = newTime
		if err := w.WritePacket(ci, packet.Data()); err != nil {
			fmt.Printf("Error writing packet: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("PCAP file rewritten successfully!")
}
