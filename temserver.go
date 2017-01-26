package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coderconvoy/templater/configmanager"
)

type Loose struct {
	FileName string
	Style    string
}

var configMan *configmanager.Manager

func staticFiles(w http.ResponseWriter, r *http.Request) {
	fPath, err := configMan.GetFilePath(r.URL.Host, r.URL.Path)

	if err != nil {
		w.Write([]byte("Bad File Request"))
	}

	http.ServeFile(w, r, fPath)
}

func bigHandler(w http.ResponseWriter, r *http.Request) {
	//Handle restyling options with a style cookie
	host := r.URL.Host
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

	fmt.Println("Path---", p)
	// Empty for index

	if p == "" {
		err = configMan.TryTemplate(w, host, "index", Loose{"", style})
		if err != nil {
			fmt.Fprintf(w, "Could not load index, err = %s", err)
			fmt.Printf("Could not load index, err = %s", err)
		}
		return
	}
	if p == "favicon.ico" {
		http.ServeFile(w, r, "/files/s/favicon.ico")
		return
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
	err = configMan.TryTemplate(w, host, p, Loose{"", style})
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

	flag.Parse()
	var err error

	configMan, err = configmanager.NewManager(*config)
	if err != nil {
		fmt.Println("config error:", err)
		return
	}

	http.HandleFunc("/s/", staticFiles)

	http.HandleFunc("/", bigHandler)

	fmt.Println("Started")

	http.ListenAndServe(":"+*port, nil)

}
