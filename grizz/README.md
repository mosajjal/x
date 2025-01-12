# Grizz, the fastest lookup


Grizz is designed to serve millions of requests per second. Its main use case is for Threat Intelligence, IP Lookups, and other similar use cases.
The design is built around the notion of "yes/no" for a request. meaning each API is mandated to have a plain-text integer response (0 or 1).

note that everything will be stored in memory, so make sure you have enough memory to serve large files.

Grizz can serve lookups through:
- [x] HTTP API (GET)
- [ ] Plain TCP
- [ ] Plain UDP
- [ ] Unix Socket
- [ ] DNS

## Supported Entities

- [x] IPv4
- [x] IPv6
- [ ] Domain
- [ ] SHA256


## Supported Input Formats

IP:

- [x] Single IP
- [x] List of IPs
- [x] List of CIDRs

## Configuration

The configuration is done through `hcl`. Here is an example configuration:

```hcl
endpoint "ip" "bad_ips" {
    modes = ["http", "socket"] # can be "http", "socket" or both
    http_listener = "tcp://:8080" # the address to listen on
    socket_listener = "tcp://:8081" # the address to listen on
    http_base_path = "/" # the base path for the API
    file = "bad_ips.txt" # can be a file or a URL. will respect HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables
    file_format = "plain" # a plain file is a list of IPv4, IPv6, CIDR, or a mix of them in a plain text file 
    inverted = false # if inverted is true, the response will be 0 for a hit, and 1 for a miss
    auto_reload = "5m" # the interval to reload the file. 0 to disable
}

endpoint "ip" "tor_exit_nodes" {
    file = "https://check.torproject.org/torbulkexitlist"
    file_format = "plain"
    inverted = false
    auto_reload = "1h"
}

endpoint "composite" "bad_ips_and_tor" {
    endpoints = ["bad_ips", "tor_exit_nodes"]
    operator = "and" # can be "and" or "or"
}
```
