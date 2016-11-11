package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_HeadedMD(t *testing.T) {
	b, err := ioutil.ReadFile("test_data/t1.md")
	if err != nil {
		t.Logf("File Not Read : %s\n", err)
	}

	m := HeadedMD(b)
	if m["title"] != "poop" {
		t.Log("tile not poop")
		t.Fail()
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
