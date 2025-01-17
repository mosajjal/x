# Gopherchain

An alternative to proxychains using Linux Network Namespaces and tun2socks

## Features

- Uses kernel network namespaces instead of LD_PRELOAD hooks
- Cannot be bypassed by statically linked binaries
- MagicDNS to automatically redirect all the DNS requests to a remote DNS server (DoT, DoH, DoQ, TCP supported)

## Requirements

- Linux kernel with network namespace support
- Root privileges
- Go 1.23+

## Quick Start

### Install

```bash
go install github.com/mosajjal/go-exp/gopherchain@latest
```

### Basic usage

```bash
sudo ./gopherchain -proxy socks5://proxy.example.com:1080

# Run any command under the namsepace
sudo nsenter --net=/run/netns/gopherchain curl ipinfo.io
```

> [!WARNING]
> if you get a DNS error, you might need to set your host DNS server to an IP other than localhost or unreachable networks
> or use the -magicdns flag 

### Use with MagicDNS
```bash
sudo ./gopherchain -proxy socks5://proxy.example.com:1080 -magicdns https://dns.adguard.com

# Run any command under the namsepace
sudo nsenter --net=/run/netns/gopherchain curl ipinfo.io
```

## Configuration

```
usage of ./gopherchain
  -device string
        TUN device name [GOPHERCHAIN_DEVICE] (default "gopherchaintun0")
  -ipmask string
        IP address of the TUN device [GOPHERCHAIN_IPMASK] (default "100.200.200.1/32")
  -loglevel string
        Log level [GOPHERCHAIN_LOGLEVEL] (default "debug")
  -magicdns string
        if a dns server value is specified,
                starts a local dns server and forwards all traffic to udp53 to this dns server
          - udp://1.1.1.1:53
          - tcp://9.9.9.9:5353
          - https://dns.adguard.com
          - quic://dns.adguard.com:8853
          - tcp-tls://dns.adguard.com:853 [GOPHERCHAIN_MAGICDNS]
  -nsname string
        Name of the new network namespace [GOPHERCHAIN_NSNAME] (default "gopherchain")
  -proxy string
        Proxy address. can't be localhost [GOPHERCHAIN_PROXY] (default "socks5://10.1.1.1:1080")
```

## Use in Systemd

After gopherchain is installed and running, you can start any systemd unit in the namespace by adding the following lines to the unit file:

```ini
[Unit]
Description=My Service
# netns.service sets up the network namespace
After=network-online.target
Requires=network-online.target

[Service]
Type=simple
# The following doesn't work, app starts but every network request fails
NetworkNamespacePath=/run/netns/gopherchain
User=root
Group=root
ExecStart=/usr/bin/app

[Install]
WantedBy=multi-user.target
```

note that the same DNS issue applies here. you might need to set the DNS server in the namespace to a reachable IP through the SOCKS proxy.
additionally, you need to make sure your SOCKS5 proxy supports handling UDP packets.

### Security Notes
- All DNS requests are routed through the proxy
- IPv6 is automatically disabled in the namespace
- Process isolation is enforced at kernel level
- Cannot be bypassed by statically linked binaries

