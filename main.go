package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/coderconvoy/dbase"
	"github.com/coderconvoy/templater/cfm"
)

type Loose struct {
	FileName string
	Style    string
}

var configMan *cfm.Manager

func staticFiles(w http.ResponseWriter, r *http.Request) {
	fPath, err := configMan.GetFilePath(r.Host, r.URL.Path)

	if err != nil {
		w.Write([]byte("Bad File Request"))
	}

	http.ServeFile(w, r, fPath)
}

func bigHandler(w http.ResponseWriter, r *http.Request) {
	//Handle restyling options with a style cookie
	host := r.Host
	styleC, cerr := r.Cookie("style")
	style := ""
	if cerr == nil {
		style = styleC.Value
	}

	s2 := r.URL.Query().Get("style")
	if s2 != "" {
		style = s2
		styleC = &http.Cookie{Name: "style", Value: style, Expires: time.Now().Add(time.Hour * 24)}
	}

	if styleC != nil {
		http.SetCookie(w, styleC)
	}

	//allow errors

	var err error
	p := strings.TrimPrefix(r.URL.Path, "/")

	cfm.LogTof(host, "Path---%s", p)
	// Empty for index

	if p == "" {
		err = configMan.TryTemplate(w, host, "index", Loose{"index.md", style})
		if err != nil {
			fmt.Fprintf(w, "Could not load index, err = %s", err)
			dbase.QLogf("Could not load index, err = %s", err)
		}
		return
	}

	//Top Level fake to s
	fp, err := configMan.GetFilePath(r.Host, "s/"+p)
	if err == nil {
		_, err2 := os.Stat(fp)
		if err2 == nil {
			http.ServeFile(w, r, fp)
			return
		}
	}
	//try template
	for k, v := range p {
		if v == '/' {
			err = configMan.TryTemplate(w, host, p[:k], Loose{p[k+1:], style})
			if err != nil {
				dbase.QLog("No template found:" + err.Error())
			}

			if err == nil {
				return
			}
			dbase.QLog(err)
		}
	}
	err = configMan.TryTemplate(w, host, p, Loose{p, style})
	if err == nil {
		return
	}

	err = configMan.TryTemplate(w, host, "loose", Loose{p + ".md", style})
	if err != nil {
		dbase.QLog(err)
		fmt.Fprintln(w, err)
	}

}

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

	http.HandleFunc("/s/", staticFiles)

	http.HandleFunc("/", bigHandler)

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
