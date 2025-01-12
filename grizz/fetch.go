package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/mosajjal/x/grizz/ip"
)

// fetch a HTTP, HTTPS, File or path and return an io.Reader
func fetchFile(file string) (io.ReadCloser, error) {
	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
	return os.Open(file)
}

// parseLineToCIDR reads string and returns a CIDR address
// basically, adding a /32 to the IPv4 and /128 to the IPv6
// and leaving the CIDR as is
// it DOES NOT try to validate the IP address.
func parseLineToCIDR(line string) string {
	if strings.Contains(line, "/") {
		return line
	}
	if strings.Contains(line, ":") {
		return fmt.Sprintf("%s/128", line)
	}
	if strings.Contains(line, ".") {
		return fmt.Sprintf("%s/32", line)
	}
	return line
}

// getIPNets reads a plaintext file containing a list of
// ipv4, ipv6, or CIDR addresses and returns a slice of *net.IPNet
func getIPNets(f io.Reader) ([]*net.IPNet, error) {
	var ipnets []*net.IPNet

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %s", err)
		}

		_, ipnet, err := net.ParseCIDR(parseLineToCIDR(record[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse CIDR %s: %s", parseLineToCIDR(record[0]), err)
		}
		ipnets = append(ipnets, ipnet)
	}
	return ipnets, nil
}

func populateLooker(f io.Reader) (*ip.Looker, error) {

	looker := ip.NewTrie()
	ipnets, err := getIPNets(f)
	if err != nil {
		return nil, err
	}
	looker.Reload(ipnets...)
	return looker, nil
}
