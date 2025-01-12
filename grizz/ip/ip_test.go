package ip

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/carlmjohnson/requests"
)

// tests and benchmarks for the ip package

var someNZRanges = []string{
	"14.1.32.0/19",
	"14.102.98.0/23",
	"14.128.4.0/22",
	"14.137.0.0/19",
	"27.96.64.0/22",
	"27.110.120.0/21",
	"27.111.12.0/22",
	"27.121.88.0/22",
	"27.121.96.0/22",
	"27.123.20.0/22",
	"27.252.0.0/16",
	"36.50.206.0/23",
	"43.224.120.0/22",
	"43.224.250.0/23",
	"43.225.200.0/22",
	"43.226.8.0/22",
	"43.226.216.0/22",
	"43.228.184.0/22",
	"43.231.192.0/22",
	"43.239.92.0/22",
	"43.239.96.0/22",
	"43.239.180.0/22",
	"43.239.250.0/24",
}

func listToLooker(l []string) *Looker {
	t := NewTrie()
	for _, v := range l {
		_, ipnet, _ := net.ParseCIDR(v)
		t.Insert(ipnet)
	}
	return t
}

func BenchmarkLookerContains(b *testing.B) {
	// Setup test data
	testIPs := []string{
		"43.239.250.1", // hit
		"43.226.216.1", // hit
		"8.8.8.8",      // miss
		"27.252.1.1",   // hit
		"192.168.1.1",  // miss
	}

	// Create looker from test data
	looker := listToLooker(someNZRanges)

	// Parse IPs once before benchmark
	ips := make([]net.IP, len(testIPs))
	for i, ipStr := range testIPs {
		ips[i] = net.ParseIP(ipStr)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := ips[i%len(ips)]
		looker.Contains(ip)
	}
}

func BenchmarkUnixSocketParallel(b *testing.B) {
	testIPs := []string{
		"43.239.250.1", // hit
		"43.226.216.1", // hit
		"8.8.8.8",      // miss
		"27.252.1.1",   // hit
		"192.168.1.1",  // miss
	}

	looker := listToLooker(someNZRanges)

	// Server startup with WaitGroup
	ipsrv := NewNetworkServer(looker, "unix:///tmp/ip.sock")
	go ipsrv.ListenAndServe()

	time.Sleep(1000 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine gets its own connection
		conn, err := net.Dial("unix", "/tmp/ip.sock")
		if err != nil {
			b.Fatal(err)
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		i := 0
		for pb.Next() {
			ip := testIPs[i%len(testIPs)]
			if _, err := conn.Write([]byte(ip + "\n")); err != nil {
				b.Fatal(err)
			}
			n, err := conn.Read(buf)
			if err != nil {
				b.Fatal(err)
			}
			if n != 2 {
				b.Fatalf("unexpected response: %s", buf[:n])
			}
			i++
		}
	})
}

func BenchmarkHTTP(b *testing.B) {
	testIPs := []string{
		"43.239.250.1", // hit
		"43.226.216.1", // hit
		"8.8.8.8",      // miss
		"27.252.1.1",   // hit
		"192.168.1.1",  // miss
	}

	looker := listToLooker(someNZRanges)

	// Server startup with WaitGroup

	ipsrv := NewHTTPServer(looker, "tcp://:10043", "/")
	go ipsrv.ListenAndServe()

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// use a HTTP Client to send requests
		i := 0
		for pb.Next() {
			var s string
			ip := testIPs[i%len(testIPs)]
			err := requests.
				URL(fmt.Sprintf("http://localhost:10043/%s", ip)).
				ToString(&s).
				Fetch(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}

	})
	// for i := 0; i < b.N; i++ {
	// 	ip := testIPs[i%len(testIPs)]
	// 	var s string
	// 	err := requests.
	// 		URL(fmt.Sprintf("http://localhost:10043/%s", ip)).
	// 		ToString(&s).
	// 		Fetch(context.Background())
	// 	if err != nil {
	// 		b.Fatal(err)
	// 	}
	// }

}
