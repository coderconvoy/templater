package main

import (
	"flag"
	"fmt"
	"github.com/coderconvoy/templater/tempower"
	"net/http"
	"strings"
	"time"
)

type Loose struct {
	FileName string
	Style    string
}

var templates *tempower.PowerTemplate

func bigHandler(w http.ResponseWriter, r *http.Request) {
	//Handle restyling options with a style cookie
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
		err = templates.ExecuteTemplate(w, "index", Loose{"", style})
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
			err = templates.ExecuteTemplate(w, p[:k], Loose{p[k+1:], style})

			if err == nil {
				return
			}
			errs = append(errs, err)
		}
	}
	err = templates.ExecuteTemplate(w, p, Loose{"", style})
	if err == nil {
		return
	}
	errs = append(errs, err)

	//Default
	fmt.Println(errs)
	err = templates.ExecuteTemplate(w, "loose", Loose{p + ".md", style})
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
	fmt.Println("ended")

}

func main() {

	root := flag.String("r", "", "root file path for access to files")
	port := flag.String("p", "8080", "port to bind to")
	flag.Parse()
	templates = tempower.NewPowerTemplate(*root+"/templates/*.html", *root)

	fs := http.FileServer(http.Dir(*root + "/s"))
	http.Handle("/s/", http.StripPrefix("/s/", fs))

	http.HandleFunc("/", bigHandler)

	fmt.Println("Started")
	http.ListenAndServe(":"+*port, nil)

}
