package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/coderconvoy/templater/blob"
	"github.com/coderconvoy/templater/tempower"
	"io"
	"net/http"
	"strings"
	"time"
)

type Loose struct {
	FileName string
	Style    string
}

var templates *tempower.PowerTemplate

func tryTemplate(w io.Writer, p string, data interface{}) error {
	b := new(bytes.Buffer)
	for i := 0; i < 10; i++ {
		err := templates.ExecuteTemplate(b, p, data)
		if err == nil {
			w.Write(b.Bytes())
			return nil
		}
		if err != blob.DeadBlob() {
			return err
		}
	}
	return fmt.Errorf("Tried too many times to access blob")
}

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
		err = tryTemplate(w, "index", Loose{"", style})
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
			err = tryTemplate(w, p[:k], Loose{p[k+1:], style})

			if err == nil {
				return
			}
			errs = append(errs, err)
		}
	}
	err = tryTemplate(w, p, Loose{"", style})
	if err == nil {
		return
	}
	errs = append(errs, err)

	//Default
	fmt.Println(errs)
	err = tryTemplate(w, "loose", Loose{p + ".md", style})
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
	fmt.Println("ended")

}

func editToTLS(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/edit/") && (r.URL.Path != "/edit") {
		bigHandler(w, r)
		return
	}

	h := r.URL.Host
	if h != "" {
		http.Redirect(w, r, "https://"+r.URL.Host+"/edit", 301)
		return
	}

	fmt.Fprintf(w, "please replace http with https in link")

}

type EditHandle struct{}

func (eh *EditHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/edit/") && (r.URL.Path != "/edit") {

		bigHandler(w, r)
		return
	}
	fmt.Fprintf(w, "Hello tls")

}

func main() {

	root := flag.String("r", "", "root file path for access to files")
	port := flag.String("p", "8080", "port to bind to")
	//tlsport := flag.String("tls", "8430", "Tls Port to bind to")
	flag.Parse()
	templates = tempower.NewPowerTemplate(*root+"/templates/*.html", *root)

	fs := http.FileServer(http.Dir(*root + "/s"))
	http.Handle("/s/", http.StripPrefix("/s/", fs))

	http.HandleFunc("/", editToTLS)

	fmt.Println("Started")
	http.ListenAndServe(":"+*port, nil)

	//err := http.ListenAndServeTLS(":"+*tlsport, "/home/matthew/keys/local.crt", "/home/matthew/keys/local.key", &EditHandle{})
	//fmt.Println(err)

}
