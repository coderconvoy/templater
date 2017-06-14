package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coderconvoy/dbase"
	"github.com/coderconvoy/templater/configmanager"
)

type Loose struct {
	FileName string
	Style    string
}

var configMan *configmanager.Manager

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

	errs := make([]error, 0)
	var err error
	p := strings.TrimPrefix(r.URL.Path, "/")

	dbase.QLog("Host---", host)
	dbase.QLog("Path---", p)
	// Empty for index

	if p == "" {
		err = configMan.TryTemplate(w, host, "index", Loose{"index.md", style})
		if err != nil {
			fmt.Fprintf(w, "Could not load index, err = %s", err)
			fmt.Printf("Could not load index, err = %s", err)
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

			if err == nil {
				return
			}
			errs = append(errs, err)
		}
	}
	err = configMan.TryTemplate(w, host, p, Loose{p, style})
	if err == nil {
		return
	}
	errs = append(errs, err)

	//Default
	fmt.Println(errs)
	err = configMan.TryTemplate(w, host, "loose", Loose{p + ".md", style})
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
	fmt.Println("ended")

}

func main() {
	config := flag.String("c", "config.json", "Path to JSON Config file")
	port := flag.String("p", "80", "Port")
	debug := flag.Bool("d", false, "Debug to stdout")
	flag.Parse()
	if *debug {
		fmt.Println("Debugging to stdout")
		dbase.SetQLogger(dbase.FmtLog{})
	}

	var err error

	configMan, err = configmanager.NewManager(*config)
	if err != nil {
		dbase.QLog("config error:", err)
		return
	}

	http.HandleFunc("/s/", staticFiles)

	http.HandleFunc("/", bigHandler)

	dbase.QLog("Started")

	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		dbase.QLog(err)
	}

}
