package cfm

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/coderconvoy/lazyf"
)

type proxySite struct {
	DomList
	*httputil.ReverseProxy
}

func (proxySite) Log(s string) {
}

func (proxySite) Update() {
}

func NewProxySite(lz lazyf.LZ, root string) (*proxySite, error) {
	url, err := url.Parse(lz.PStringD("", "proxy", "Proxy"))
	if err != nil {
		return nil, err
	}
	return &proxySite{DomList: lz.PStringAr("host", "Host"),
		//httputil.NewSingleHostReverseProxy(url),
		ReverseProxy: &httputil.ReverseProxy{Director: proxyDirector(url)},
	}, nil
}

func proxyDirector(target *url.URL) func(req *http.Request) {
	return func(req *http.Request) {

		fmt.Println("Redirecting:", req.URL)

		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		//req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

	}

}
