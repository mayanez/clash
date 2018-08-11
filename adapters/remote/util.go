package adapters

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	C "github.com/Dreamacro/clash/constant"
)

// DelayTest get the delay for the specified URL
func DelayTest(proxy C.Proxy, url string) (t int16, err error) {
	addr, err := urlToAddr(url)
	if err != nil {
		return
	}

	start := time.Now()
	instance, err := proxy.Generator(&addr)
	if err != nil {
		return
	}
	defer instance.Close()
	transport := &http.Transport{
		Dial: func(string, string) (net.Conn, error) {
			return instance.Conn(), nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := http.Client{Transport: transport}
	req, err := client.Get(url)
	if err != nil {
		return
	}
	req.Body.Close()
	t = int16(time.Since(start) / time.Millisecond)
	return
}

func urlToAddr(rawURL string) (addr C.Addr, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else if u.Scheme == "http" {
			port = "80"
		} else {
			err = fmt.Errorf("%s scheme not Support", rawURL)
			return
		}
	}

	addr = C.Addr{
		AddrType: C.AtypDomainName,
		Host:     u.Hostname(),
		IP:       nil,
		Port:     port,
	}
	return
}

func selectFast(in chan interface{}) chan interface{} {
	out := make(chan interface{})
	go func() {
		p, open := <-in
		if open {
			out <- p
		}
		close(out)
		for range in {
		}
	}()

	return out
}