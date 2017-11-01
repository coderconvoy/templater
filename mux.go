package main

import (
	"net/http"
	"strings"

	"github.com/coderconvoy/templater/cfm"
)

type SafeMux struct{}

func (s SafeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/s/") {
		staticFiles(w, r)
		return
	}
	bigHandler(w, r)
}

type InsecMux struct {
	secDoms []string
}

func (s InsecMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fnd := ""
	cfm.Logq("Insec:" + r.Host)
	host := strings.Split(r.Host, ":")[0]
	for _, v := range s.secDoms {
		if v == host {
			fnd = v
			break
		}
	}
	if fnd != "" {
		http.Redirect(w, r, "https://"+host+r.URL.Path, 302)
		return
	}
	SafeMux{}.ServeHTTP(w, r)
}
