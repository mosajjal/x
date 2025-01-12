package ip

import (
	"net"

	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	l        *Looker
	addr     string
	basepath string
}

func NewFibreServer(l *Looker, addr string, basepath string) *FiberServer {
	return &FiberServer{l: l, addr: addr, basepath: basepath}
}

func (s *FiberServer) ListenAndServe() error {
	app := fiber.New(fiber.Config{
		Prefork: true,
	})

	app.Get(s.basepath+":ip", func(c *fiber.Ctx) error {
		ipStr := c.Params("ip")
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return c.SendString("0")
		}
		if s.l.Contains(ip) {
			return c.SendString("1")
		}
		return c.SendString("0")
	})
	return app.Listen(s.addr)
}
