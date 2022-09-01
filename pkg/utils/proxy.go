package utils

import (
	"golang.org/x/net/proxy"
	"net/url"
)

func GetDial(p string) proxy.ContextDialer {
	u, err := url.Parse(p)
	if err != nil {
		return proxy.Direct
	}
	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return proxy.Direct
	}

	d, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return proxy.Direct
	}
	return d
}
