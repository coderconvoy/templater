package main

import (
	"net/http"
	"strings"

	"github.com/coderconvoy/templater/cfm"
)

type SafeMux struct{}

func (s SafeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cf, err := configMan.GetConfig(r.Host)
	if err != nil {
		http.Error(w, "No Site under that domain type", 404)
		return
	}
	cf.ServeHTTP(w, r)
}

type InsecMux struct {
	secDoms map[string]bool
	secPort string
}

func (s InsecMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cfm.Logq("Insec:" + r.Host)
	sp := strings.Split(r.Host, ":")
	host := sp[0]
	port := ""
	if s.secPort != "443" {
		port = ":" + s.secPort
	}

	trydomain := func(dom string) bool {
		_, ok := s.secDoms[dom]
		if ok {
			http.Redirect(w, r, "https://"+dom+port+r.URL.Path, 302)
			return true
		}
		return false
	}
	if trydomain(host) {
		return
	}
	if trydomain("www." + host) {
		return
	}
	if trydomain(strings.TrimPrefix(host, "www.")) {
		return
	}

	SafeMux{}.ServeHTTP(w, r)
}
