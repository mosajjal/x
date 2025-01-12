package ip

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/evanphx/wildcat"
	"github.com/panjf2000/gnet/v2"
)

// HTTPServer is a plain HTTP server
// it receives a HTTP request and returns a response
// to create a new HTTP server, use NewHTTPServer
type HTTPServer struct {
	gnet.BuiltinEventEngine

	l        *Looker
	addr     string
	basepath string
}

type httpCodec struct {
	parser        *wildcat.HTTPParser
	contentLength int
	buf           []byte
}

var (
	crlf      = []byte("\r\n\r\n")
	lastChunk = []byte("0\r\n\r\n")
)

func (hc *httpCodec) parse(data []byte) (int, string, error) {

	splits := strings.Split(string(data), " ")
	if len(splits) < 2 {
		return 0, "", errors.New("invalid http request")
	}

	return 0, splits[1], nil
}

var contentLengthKey = []byte("Content-Length")

func (hc *httpCodec) getContentLength() int {
	if hc.contentLength != -1 {
		return hc.contentLength
	}

	val := hc.parser.FindHeader(contentLengthKey)
	if val != nil {
		i, err := strconv.ParseInt(string(val), 10, 0)
		if err == nil {
			hc.contentLength = int(i)
		}
	}

	return hc.contentLength
}

func (hc *httpCodec) resetParser() {
	hc.contentLength = -1
}

func (hc *httpCodec) reset() {
	hc.resetParser()
	hc.buf = hc.buf[:0]
}

func writeResponse(hc *httpCodec, body []byte) {
	// You may want to determine the URL path and write the corresponding response.
	// ...

	hc.buf = append(hc.buf, "HTTP/1.1 200 OK\r\nServer: gnet\r\nContent-Type: text/plain\r\nDate: "...)
	hc.buf = time.Now().AppendFormat(hc.buf, "Mon, 02 Jan 2006 15:04:05 GMT")
	hc.buf = append(hc.buf, "\r\nContent-Length: "...)
	hc.buf = append(hc.buf, strconv.Itoa(len(body))...)
	hc.buf = append(hc.buf, "\r\n\r\n"...)
	hc.buf = append(hc.buf, body...)
}

// OnBoot is called when the server is started
func (hs *HTTPServer) OnBoot(_ gnet.Engine) gnet.Action {
	log.Printf("http server started. listening on %s\n", hs.addr)
	return gnet.None
}

// OnOpen is called when a new connection is opened
func (hs *HTTPServer) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(&httpCodec{parser: wildcat.NewHTTPParser()})
	return nil, gnet.None
}

// OnTraffic is called when data is received
func (hs *HTTPServer) OnTraffic(c gnet.Conn) gnet.Action {
	hc := c.Context().(*httpCodec)
	buf, err := c.Peek(-1)
	if err != nil {
		fmt.Println("peek error:", err)
	}
	n := len(buf)
	// we expect the buffer to start with
	// GET /basepath/[ip] HTTP/1.1\r\n
	// since this is a HTTP 1.1 server ONLY

	nextOffset, url, err := hc.parse(buf)
	// trim the url and remove the basepath to get the IP
	ipStr := strings.TrimPrefix(url, hs.basepath)
	// parse the IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		c.Write([]byte("error: invalid IP format\n"))
		return gnet.None
	}
	body := []byte("0\n")
	if hs.l.Contains(ip) {
		body = []byte("1\n")
	}

	hc.resetParser()
	if err != nil {
		goto response
	}
	if len(buf) < nextOffset {
		goto response
	}
	writeResponse(hc, body)

response:
	if len(hc.buf) > 0 {
		c.Write(hc.buf)
	}
	hc.reset()
	c.Discard(n - len(buf))
	return gnet.None
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(l *Looker, addr string, basepath string) *HTTPServer {
	return &HTTPServer{l: l, addr: addr, basepath: basepath}
}

// ListenAndServe starts the server
func (hs *HTTPServer) ListenAndServe() error {
	return gnet.Run(hs, hs.addr, gnet.WithMulticore(true))
}
