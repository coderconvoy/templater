package main

import (
	"crypto/tls"
	"crypto/x509"
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
	secPort := lazyf.FlagString("sp", "443", "secport", "Port for TLS")

	confs, cfname := lazyf.FlagLoad("c", "config.lz")

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

	insecMux := InsecMux{
		secDoms: make(map[string]bool),
		secPort: *secPort,
	}
	for _, v := range doms {
		cert, err := tls.LoadX509KeyPair(path.Join(*keyloc, v, pubkeyf), path.Join(*keyloc, v, privkeyf))
		if err != nil {
			cfm.Logf("--X--%s\n", v)
			continue
		}
		//get domains out of certificate
		xcert, err := x509.ParseCertificate(cert.Certificate[0])

		insecMux.secDoms[xcert.Subject.CommonName] = true
		for _, d := range xcert.DNSNames {
			insecMux.secDoms[d] = true
		}

		scfg.Certificates = append(scfg.Certificates, cert)
	}

	scfg.BuildNameToCertificate()

	tlserver := http.Server{
		Addr:      ":" + *secPort,
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
