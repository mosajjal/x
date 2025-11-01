// Package main provides fakedns, a programmable DNS server that returns static IPs
// for configured domain patterns. It supports prefix, suffix, and FQDN matching with
// upstream DNS fallback for unmatched domains.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-collections/collections/tst"
	"github.com/miekg/dns"
	"github.com/mosajjal/dnsclient"
)

// FakeDNS represents the DNS server configuration and routing tables.
type FakeDNS struct {
	dnsclient.Client
	routePrefixes *tst.TernarySearchTree // TST for prefix matching
	routeSuffixes *tst.TernarySearchTree // TST for suffix matching (reversed strings)
	routeFQDNs    map[string]TreeValue   // Map for exact FQDN matching
	UDPPort       uint64
	TCPPort       uint64
}

var (
	fakeDNS FakeDNS
	dnsLock sync.RWMutex
	logger  *slog.Logger
)

const (
	// Match types for domain routing
	matchPrefix uint8 = 1
	matchSuffix uint8 = 2
	matchFQDN   uint8 = 3
)

// TreeValue is inserted into TSTs as value for each prefix, suffix and FQDN.
type TreeValue struct {
	Entry string  // The original domain pattern
	Mode  uint8   // Match type (prefix, suffix, or FQDN)
	IP    net.IP  // IP address to return for this pattern
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// inDomainList returns true if the domain is meant to be SKIPPED and not go through sni proxy
func (c FakeDNS) inDomainList(fqdn string) bool {
	fqdnLower := strings.ToLower(fqdn)
	// check for fqdn match
	if c.routeFQDNs[fqdnLower].Mode == matchFQDN {
		return false
	}
	// check for prefix match
	if longestPrefix := c.routePrefixes.GetLongestPrefix(fqdnLower); longestPrefix != nil {
		// check if the longest prefix is present in the type hashtable as a prefix
		if c.routeFQDNs[longestPrefix.(TreeValue).Entry].Mode == matchPrefix {
			return false
		}
	}
	// check for suffix match. Note that suffix is just prefix reversed
	if longestSuffix := c.routeSuffixes.GetLongestPrefix(reverse(fqdnLower)); longestSuffix != nil {
		// check if the longest suffix is present in the type hashtable as a suffix
		if c.routeFQDNs[longestSuffix.(TreeValue).Entry].Mode == matchSuffix {
			return false
		}
	}
	return true
}

// ApproperiateIP returns the approperiate IP for the given FQDN
func (c FakeDNS) ApproperiateIP(fqdn string) net.IP {
	fqdnLower := strings.ToLower(fqdn)
	// check for fqdn match
	if c.routeFQDNs[fqdnLower].Mode == matchFQDN {
		return c.routeFQDNs[fqdnLower].IP
	}
	// check for prefix match
	if longestPrefix := c.routePrefixes.GetLongestPrefix(fqdnLower); longestPrefix != nil {
		// check if the longest prefix is present in the type hashtable as a prefix
		if c.routeFQDNs[longestPrefix.(TreeValue).Entry].Mode == matchPrefix {
			return c.routeFQDNs[longestPrefix.(TreeValue).Entry].IP
		}
	}
	// check for suffix match. Note that suffix is just prefix reversed
	if longestSuffix := c.routeSuffixes.GetLongestPrefix(reverse(fqdnLower)); longestSuffix != nil {
		// check if the longest suffix is present in the type hashtable as a suffix
		if c.routeFQDNs[longestSuffix.(TreeValue).Entry].Mode == matchSuffix {
			return c.routeFQDNs[longestSuffix.(TreeValue).Entry].IP
		}
	}
	return nil
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// LoadDomainsCsv loads a domains Csv file/URL. returns 3 parameters:
// 1. a TST for all the prefixes (type 1)
// 2. a TST for all the suffixes (type 2)
// 3. a hashtable for all the full match fqdn (type 3)
func (c *FakeDNS) LoadDomainsCsv(Filename string) error {
	c.routePrefixes = tst.New()
	c.routeSuffixes = tst.New()
	c.routeFQDNs = make(map[string]TreeValue)

	logger.Info("Loading the domain from file/url")
	var scanner *bufio.Scanner
	if strings.HasPrefix(Filename, "http://") || strings.HasPrefix(Filename, "https://") {
		logger.Info("domain list is a URL, trying to fetch")
		client := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := client.Get(Filename)
		if err != nil {
			logger.Error("", err)
			return err
		}
		logger.Info("(re)fetching URL", "url", Filename)
		defer resp.Body.Close()
		scanner = bufio.NewScanner(resp.Body)

	} else {
		file, err := os.Open(Filename)
		if err != nil {
			return err
		}
		logger.Info("(re)loading File", "file", Filename)
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		lowerCaseLine := strings.ToLower(scanner.Text())
		// split the line by comma to understand thelogger.c
		fqdn := strings.Split(lowerCaseLine, ",")
		if len(fqdn) != 3 {
			logger.Info(lowerCaseLine + " is not a valid line, assuming FQDN")
			fqdn = []string{lowerCaseLine, "fqdn"}
		}
		// add the fqdn to the hashtable with its type
		switch entryType := fqdn[1]; entryType {
		case "prefix":
			c.routeFQDNs[fqdn[0]] = TreeValue{Mode: matchPrefix, IP: net.ParseIP(fqdn[2]), Entry: fqdn[0]}
			c.routePrefixes.Insert(fqdn[0], c.routeFQDNs[fqdn[0]])
		case "suffix":
			c.routeFQDNs[fqdn[0]] = TreeValue{Mode: matchSuffix, IP: net.ParseIP(fqdn[2]), Entry: fqdn[0]}
			// suffix match is much faster if we reverse the strings and match for prefix
			c.routeSuffixes.Insert(reverse(fqdn[0]), c.routeFQDNs[fqdn[0]])
		case "fqdn":
			c.routeFQDNs[fqdn[0]] = TreeValue{Mode: matchFQDN, IP: net.ParseIP(fqdn[2]), Entry: fqdn[0]}
		default:
			//logger.Warnf("%s is not a valid line, assuming fqdn", lowerCaseLine)
			logger.Info(lowerCaseLine + " is not a valid line, assuming FQDN")
			c.routeFQDNs[fqdn[0]] = TreeValue{Mode: matchFQDN, IP: net.ParseIP(fqdn[2]), Entry: fqdn[0]}
		}
	}
	logger.Info(fmt.Sprintf("%s loaded with %d prefix, %d suffix and %d fqdn", Filename, c.routePrefixes.Len(), c.routeSuffixes.Len(), len(c.routeFQDNs)-c.routePrefixes.Len()-c.routeSuffixes.Len()))

	return nil
}

func (c *FakeDNS) performExternalAQuery(fqdn string) ([]dns.RR, time.Duration, error) {
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	msg := dns.Msg{}
	msg.RecursionDesired = true
	msg.SetQuestion(fqdn, dns.TypeA)
	msg.SetEdns0(1232, true)
	dnsLock.Lock()
	if c.Client == nil {
		return nil, 0, fmt.Errorf("DNS client is not initialised")
	}
	res, trr, err := c.Client.Query(context.Background(), &msg)
	if err != nil {
		if err.Error() == "EOF" {
			logger.Info("reconnecting DNS...")
			// dnsc.C.Close()
			// dnsc.C, err = dnsclient.New(c.UpstreamDNS, true)
			err = c.Client.Reconnect()
		}
	}
	dnsLock.Unlock()
	return res, trr, err
}

func (c FakeDNS) processQuestion(q dns.Question) ([]dns.RR, error) {
	if !c.inDomainList(q.Name) {
		// Return the public IP.
		rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, c.ApproperiateIP(q.Name)))
		if err != nil {
			return nil, err
		}

		logger.Info("returned sniproxy address for domain", "fqdn", q.Name)

		return []dns.RR{rr}, nil
	}

	// Otherwise do an upstream query and use that answer.
	resp, rtt, err := c.performExternalAQuery(q.Name)
	if err != nil {
		return nil, err
	}

	logger.Info("[DNS] returned origin address", "fqdn", q.Name, "rtt", rtt)

	return resp, nil
}

func (c FakeDNS) lookupDomain4(domain string) (net.IP, error) {
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	rAddrDNS, _, err := c.performExternalAQuery(domain)
	if err != nil {
		return nil, err
	}
	if len(rAddrDNS) > 0 {
		if rAddrDNS[0].Header().Rrtype == dns.TypeCNAME {
			return c.lookupDomain4(rAddrDNS[0].(*dns.CNAME).Target)
		}
		if rAddrDNS[0].Header().Rrtype == dns.TypeA {
			return rAddrDNS[0].(*dns.A).A, nil
		}
	} else {
		return nil, fmt.Errorf("[DNS] Empty DNS response for %s", domain)
	}
	return nil, fmt.Errorf("[DNS] Unknown type %s", dns.TypeToString[rAddrDNS[0].Header().Rrtype])
}

func handleDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode != dns.OpcodeQuery {
		m.SetRcode(r, dns.RcodeNotImplemented)
		w.WriteMsg(m)
		return
	}

	for _, q := range m.Question {
		answers, err := fakeDNS.processQuestion(q)
		if err != nil {
			logger.Error("", err)
			continue
		}
		m.Answer = append(m.Answer, answers...)
	}

	w.WriteMsg(m)
}

func main() {
	dns.HandleFunc(".", handleDNS)
	// start DNS UDP serverUdp

	// set up flags
	flag.Uint64Var(&fakeDNS.UDPPort, "udp", 53, "UDP port to listen on, 0 will disable UDP")
	flag.Uint64Var(&fakeDNS.TCPPort, "tcp", 0, "TCP port to listen on, 0 will disable TCP")
	// upstream DNS
	upstreamDNS := flag.String("upstream", "udp://1.0.0.1:53", "Upstream DNS server to use, example: udp://1.1.1.1:53")
	ruleFile := flag.String("rule", "", "Rule file to use, example: /etc/sniproxy/rule.list")

	flag.Parse()

	// set up upstream DNS
	dnsClient, err := dnsclient.New(*upstreamDNS, true, "")
	if err != nil {
		logger.Error("Failed to create DNS client", err)
		panic(1)
	}
	fakeDNS.Client = dnsClient

	// set up rule
	if *ruleFile != "" {
		err = fakeDNS.LoadDomainsCsv(*ruleFile)
		if err != nil {
			logger.Error("Failed to load rule", err)
			panic(1)
		}
	}
	if fakeDNS.UDPPort != 0 {
		go func() {
			serverUDP := &dns.Server{Addr: fmt.Sprintf(":%d", fakeDNS.UDPPort), Net: "udp"}
			logger.Info("Started UDP DNS", "host", "0.0.0.0", "port", fakeDNS.UDPPort)
			err := serverUDP.ListenAndServe()
			defer serverUDP.Shutdown()
			if err != nil {
				logger.Error("Error starting UDP DNS server", err)
				logger.Info(fmt.Sprintf("Failed to start server: %s\nYou can run the following command to pinpoint which process is listening on port %d\nsudo ss -pltun -at '( dport = :%d or sport = :%d )'", err.Error(), fakeDNS.TCPPort, fakeDNS.TCPPort, fakeDNS.TCPPort))
				panic(2)
			}
		}()
	}
	// start DNS UDP serverTcp
	if fakeDNS.TCPPort != 0 {
		go func() {
			serverTCP := &dns.Server{Addr: fmt.Sprintf(":%d", fakeDNS.TCPPort), Net: "tcp"}
			logger.Info("Started TCP DNS", "host", "0.0.0.0", "port", fakeDNS.TCPPort)
			err := serverTCP.ListenAndServe()
			defer serverTCP.Shutdown()
			if err != nil {
				logger.Error("Failed to start server", err)
				logger.Info(fmt.Sprintf("You can run the following command to pinpoint which process is listening on port %d\nsudo ss -pltun -at '( dport = :%d or sport = :%d )'", fakeDNS.TCPPort, fakeDNS.TCPPort, fakeDNS.TCPPort))
			}
		}()
	}

	// start DNS UDP serverTls
	// if c.BindDNSOverTLS {
	// 	go func() {
	// 		crt, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey)
	// 		if err != nil {
	// 			logger.Error("", err)
	// 			panic(2)

	// 		}
	// 		tlsConfig := &tls.Config{}
	// 		tlsConfig.Certificates = []tls.Certificate{crt}

	// 		serverTLS := &dns.Server{Addr: ":853", Net: "tcp-tls", TLSConfig: tlsConfig}
	// 		logger.Info("Started DoT DNS", "host", "0.0.0.0", "port", 853)
	// 		err = serverTLS.ListenAndServe()
	// 		defer serverTLS.Shutdown()
	// 		if err != nil {
	// 			logger.Error("", err)
	// 		}
	// 	}()
	// }

	// if c.BindDNSOverQuic {

	// 	crt, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey)
	// 	if err != nil {
	// 		logger.Error("", err)
	// 	}
	// 	tlsConfig := &tls.Config{}
	// 	tlsConfig.Certificates = []tls.Certificate{crt}

	// 	// Create the QUIC listener
	// 	doqServer, err := doqserver.New(":8853", crt, "127.0.0.1:53", true)
	// 	if err != nil {
	// 		logger.Error("", err)
	// 	}

	// 	// Accept QUIC connections
	// 	logger.Info("Starting QUIC listener on :8853")
	// 	go doqServer.Listen()

	// }
	select {}
}
