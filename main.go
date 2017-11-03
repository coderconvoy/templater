package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/coderconvoy/templater/cfm"
)

var configMan *cfm.Manager

func main() {
	config := flag.String("c", "config.json", "Path to JSON Config file")
	port := flag.String("p", "80", "Port")
	debug := flag.Bool("d", false, "Debug to stdout")
	flag.Parse()

	var err error

	configMan, err = cfm.NewManager(*config)
	if err != nil {
		cfm.Logq("config error:", err)
		return
	}

	if !*debug {
		fmt.Println("Debugging to log folders")
		cfm.SetLogger(configMan)
	} else {
		fmt.Println("Debugging to stdout")
	}

	keyLoc := configMan.KeyLoc()

	//if no keys
	if keyLoc == "" {
		cfm.Logq("Starting with no TLS")
		err := http.ListenAndServe(":"+*port, SafeMux{})
		if err != nil {
			fmt.Println(err)
			cfm.Logq(err)
		}
		return
	}

	//TLS bit could be complicated

	scfg := &tls.Config{}
	doms := configMan.Domains()

	pubkeyf := configMan.Confs().PStringD("fullchain.pem", "pubkey")
	privkeyf := configMan.Confs().PStringD("privkey.pem", "privkey")
	cfm.Logf("Keylocs:%s:%s", pubkeyf, privkeyf)

	insecMux := InsecMux{}
	for _, v := range doms {
		cert, err := tls.LoadX509KeyPair(path.Join(keyLoc, v, pubkeyf), path.Join(keyLoc, v, privkeyf))
		if err != nil {
			cfm.Logf("--X--%s\n", v)
			continue
		}
		insecMux.secDoms = append(insecMux.secDoms, v)

		scfg.Certificates = append(scfg.Certificates, cert)
	}

	scfg.BuildNameToCertificate()

	tlserver := http.Server{
		Addr:      ":443",
		Handler:   SafeMux{},
		TLSConfig: scfg,
	}

	cfm.Logq("Started")

	go func() {
		err := http.ListenAndServe(":"+*port, insecMux)
		if err != nil {
			fmt.Println(err)
			cfm.Logq(err)
		}
	}()

	log.Fatal(tlserver.ListenAndServeTLS("", ""))
}
