package cmd

import (
	"context"
	"net"
	"time"
)

func lookupIP(addr string) (string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	ip, err := r.LookupHost(context.Background(), addr)
	if err != nil {
		return "", err
	}
	return ip[0], nil
}
