package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/coderconvoy/lazyf"
	"github.com/coderconvoy/templater/cfm"
)

var configMan *cfm.Manager

func main() {
	port := lazyf.FlagString("p", "80", "port", "Port")
	debug := lazyf.FlagBool("d", "debug", "Debug to stdout")
	keyloc := lazyf.FlagString("cloc", "", "certloc", "Location of Certificate Files")

	confs, cfname := lazyf.FlagLoad("c", "config.json")

	var err error

	configMan, err = cfm.NewManager(confs, cfname)
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

	//if no keys
	if *keyloc == "" {
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
		cert, err := tls.LoadX509KeyPair(path.Join(*keyloc, v, pubkeyf), path.Join(*keyloc, v, privkeyf))
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
