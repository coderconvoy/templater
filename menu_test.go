package templater

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_Stuff(t *testing.T) {
	ar := []string{"Hello", "{", "poo:pee", "Grow:Gree", "No", "}", "Goodbye:adios"}

	m, err := NewMenu(ar)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(m)
	r, _ := json.Marshal(m.Children)
	fmt.Println(string(r))

	tree := TagTree(m.Children)
	fmt.Println(tree.String())
}

func Test_fails(t *testing.T) {
	ar := [][]string{
		[]string{"{"},
		[]string{"hello,", "{", "{"},
	}
	for _, v := range ar {
		r, err := NewMenu(v)
		fmt.Println(r)
		if err == nil {
			t.Fail()
		}
	}
}
