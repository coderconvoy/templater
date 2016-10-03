package templater

import (
	"fmt"
	"net/http"
	"testing"
)

type rr string

func (self rr) Header() http.Header {
	return nil
}

func (self rr) WriteHeader(int) {
}

func (self rr) Write(b []byte) (int, error) {
	fmt.Println(string(b))
	return len(b), nil
}

func Test_Writer(t *testing.T) {
	var w rr
	req, _ := http.NewRequest("GET", "http://poo.com/share/js/go_menu.js", nil)
	ServeSharedFile(w, req)

}
