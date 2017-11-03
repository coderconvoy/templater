package cfm

import (
	"net/http"
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

//
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
