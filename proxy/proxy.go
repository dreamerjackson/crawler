package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
)

type ProxyFunc func(*http.Request) (*url.URL, error)

type roundRobinSwitcher struct {
	proxyURLs []*url.URL
	index     uint32
}

func (r *roundRobinSwitcher) GetProxy(pr *http.Request) (*url.URL, error) {
	index := atomic.AddUint32(&r.index, 1) - 1
	u := r.proxyURLs[index%uint32(len(r.proxyURLs))]
	return u, nil
}

// RoundRobinProxySwitcher creates a proxy switcher function which rotates
// ProxyURLs on every request.
// The proxy type is determined by the URL scheme. "http", "https"
// and "socks5" are supported. If the scheme is empty,
// "http" is assumed.
func RoundRobinProxySwitcher(ProxyURLs ...string) (ProxyFunc, error) {
	if len(ProxyURLs) < 1 {
		return nil, errors.New("Proxy URL list is empty")
	}
	urls := make([]*url.URL, len(ProxyURLs))
	for i, u := range ProxyURLs {
		parsedU, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		urls[i] = parsedU
	}
	return (&roundRobinSwitcher{urls, 0}).GetProxy, nil
}
