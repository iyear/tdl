package utils

import (
	"net/url"

	"golang.org/x/net/proxy"
)

type _proxy struct{}

var Proxy = _proxy{}

func (p _proxy) GetDial(_url string) proxy.ContextDialer {
	u, err := url.Parse(_url)
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
