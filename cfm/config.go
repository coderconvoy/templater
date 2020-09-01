package cfm

import (
	"errors"
	"net/http"

	"github.com/coderconvoy/lazyf"
)

type ConfigItem interface {
	Domains() DomList
	CanHost(string) bool
	Update()
	Log(string)
	http.Handler
}

type DomList []string

func (d DomList) Domains() DomList {
	return d
}

func (d DomList) CanHost(u string) bool {
	for _, v := range d {
		if v == u {
			return true
		}

		//try after every dot for compare
		for k, c := range u {
			if c == '.' {
				if v == u[k+1:] {
					return true
				}

			}
		}
		if v == "default" {
			return true
		}
	}
	return false
}

func NewConfigItem(lz lazyf.LZ, root string) (ConfigItem, error) {
	if _, err := lz.PString("proxy", "Proxy"); err == nil {
		c, err := NewProxySite(lz, root)
		return c, err
	}

	if _, err := lz.PString("static", "Static"); err == nil {
		c, err := NewStaticSite(lz, root)
		return c, err
	}

	if _, err := lz.PString("folder", "Folder"); err == nil {
		return NewTemplateSite(lz, root)
	}

	return nil, errors.New("No ConfigItem Available")
}
