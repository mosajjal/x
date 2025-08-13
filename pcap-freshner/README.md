# PCAP Timestamp Freshener

A simple command-line tool to re-timestamp PCAP and PCAPNG files.

## Description

This tool adjusts the timestamps of all packets in a PCAP or PCAPNG file. It can shift the timestamps by a specified duration or set the latest packet's timestamp to the current time, preserving the relative time differences between all packets. This is useful for making old capture files appear "fresh" for systems that index or analyze them based on time.

## Usage

### Build

```bash
go build
```

### Run

```bash
./pcap-freshner -i <input_file> -o <output_file> [-t <time_shift>]
```

#### Flags

* `-i`: **(Required)** The input PCAP or PCAPNG file.
* `-o`: **(Required)** The output PCAP or PCAPNG file.
* `-t`: (Optional) The time shift duration. This can be a duration string like '24h', '1h30m', or the special value 'now'. If not provided, it defaults to 'now'.

## Examples

### Set Latest Packet to Current Time

This will read `old.pcap`, adjust the timestamps so that the last packet is at the current time, and write the result to `new.pcap`.

```bash
./pcap-freshner -i old.pcap -o new.pcap -t now
```

### Shift Timestamps by a Specific Duration

This will read `old.pcap`, add 24 hours to the timestamp of every packet, and write the result to `new.pcap`.

```bash
./pcap-freshner -i old.pcap -o new.pcap -t 24h
```
