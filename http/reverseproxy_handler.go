package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func forwardHandler(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	rewriteFunc := func(pr *httputil.ProxyRequest) {
		// Use received X-Forwarded-For header. This can be used if you
		// have multiple proxies in a chain. Only do this if you get
		// request from a trusted IP source.
		pr.Out.Header["X-Forwarded-For"] = pr.In.Header["X-Forwarded-For"]
		pr.SetXForwarded()

		pr.SetURL(target)
		// Keep the Host header of the inbound request
		pr.Out.Host = pr.In.Host
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: rewriteFunc,
	}

	return proxy, nil
}

// use HTTP/1.1 on upstream if Upgrade header is used on request
func forwardCustomTransportHandler(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	rewriteFunc := func(pr *httputil.ProxyRequest) {
		pr.SetXForwarded()
		pr.SetURL(target)
	}

	http11Transport := http.DefaultTransport.(*http.Transport).Clone()
	http11Transport.ForceAttemptHTTP2 = false
	http11Transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	// add other settings. e.g. ignore TLS validation
	// http11Transport.TLSClientConfig = &tls.Config{
	// 	InsecureSkipVerify: true,
	// }

	http11Upstream := &httputil.ReverseProxy{
		Rewrite:   rewriteFunc,
		Transport: http11Transport,
	}

	defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
	defaultUpstream := &httputil.ReverseProxy{
		Rewrite:   rewriteFunc,
		Transport: defaultTransport,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Upgrade is only supported by HTTP/1.1
		if r.Proto == "HTTP/1.1" && r.Header.Get("Upgrade") != "" {
			http11Upstream.ServeHTTP(w, r)
		} else {
			defaultUpstream.ServeHTTP(w, r)
		}
	}), nil
}
