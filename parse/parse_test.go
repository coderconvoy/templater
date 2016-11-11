package parse

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"testing"
)

func Test_HeadedMD(t *testing.T) {
	b, err := ioutil.ReadFile("test_data/t1.md")
	if err != nil {
		t.Logf("File Not Read : %s\n", err)
		t.Fail()
	}
	m := HeadedMD(b)
	expected := map[string]string{
		"title":    "poop",
		"date":     "23/4/2012",
		"head":     "Poop is nice to eat\nAnd so are fish\n",
		"contents": string(blackfriday.MarkdownCommon([]byte("Hi\n===\n"))),
	}
	for k, v := range expected {
		if m[k] != v {
			t.Logf("%s: expected:%s,got:%s\n", k, v, m[k])
			t.Fail()
		}
	}

}

func Test_HeadOnly(t *testing.T) {
	f, err := os.Open("test_data/t1.md")
	if err != nil {
		t.Logf("File Not Read : %s\n", err)
		t.Fail()
	}
	m, lc := HeadOnly(f)
	if lc != 6 {
		t.Logf("Line count not 6 but %d", lc)
		t.Fail()
	}
	expected := map[string]string{
		"title": "poop",
		"date":  "23/4/2012",
		"head":  "Poop is nice to eat\nAnd so are fish\n",
	}
	for k, v := range expected {
		if m[k] != v {
			t.Logf("%s: expected:%s,got:%s\n", k, v, m[k])
			t.Fail()
		}
	}

}

func test_Stuff(t *testing.T) {
	ar := []string{"Hello", "{", "poo:pee", "Grow:Gree", "No", "}", "Goodbye:adios"}

	m, err := NewMenu(ar)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(m)
	r, _ := json.Marshal(m.Children)
	fmt.Println(string(r))

	tree := TagTree(m.Children, "poo")
	fmt.Println(tree.String())
}

func Test_fails(t *testing.T) {
	ar := [][]string{
		[]string{"{"},
		[]string{"hello,", "{", "{"},
	}
	for _, v := range ar {
		r, err := NewMenu(v)
		t.Log(r)
		if err == nil {
			t.Fail()
		}
	}
}
