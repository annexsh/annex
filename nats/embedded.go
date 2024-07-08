package nats

import (
	"net"
	"strconv"

	"github.com/nats-io/nats-server/v2/server"
)

func NewEmbeddedNatsServer(hostPort string) (*server.Server, error) {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	defaultServerOpts := &server.Options{
		Host:                  host,
		Port:                  port,
		NoLog:                 true,
		NoSigs:                true,
		MaxControlLine:        4096,
		DisableShortFirstPing: true,
	}
	return server.NewServer(defaultServerOpts)
}
