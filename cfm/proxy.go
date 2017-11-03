package cfm

import (
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
		ReverseProxy: httputil.NewSingleHostReverseProxy(url),
	}, nil
}
