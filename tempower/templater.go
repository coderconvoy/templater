package tempower

import (
	"errors"
	"fmt"
	"github.com/coderconvoy/templater/blob"
	"github.com/coderconvoy/templater/parse"
	"github.com/russross/blackfriday"
	"io"
	"math/rand"
	"text/template"
)

type PowerTemplate struct {
	*template.Template
	killer func()
}

/*
   For use inside templates: Converts text sent from one template to another into a map whcih can then be accessed by {{ index }}
*/
func tDict(items ...interface{}) (map[string]interface{}, error) {
	//Throw error if not given an even number of args. This will cause an error at exec, not at parse
	if len(items)%2 != 0 {
		return nil, errors.New("tDict requires even number of arguments")
	}

	res := make(map[string]interface{}, len(items)/2)
	//Loop through args by 2, and at 1 as key, and 2 as value
	for i := 0; i < len(items)-1; i += 2 {
		k, ok := items[i].(string)
		if !ok {
			return nil, errors.New("tDict keys must be strings")
		}
		res[k] = items[i+1]
	}
	return res, nil
}

func boolSelect(cond bool, a, b interface{}) interface{} {
	if cond {
		return a
	}
	return b
}

func RandRange(l, h int) int {
	return rand.Intn(h-l) + l
}

func mdParse(d interface{}) string {
	switch v := d.(type) {
	case string:
		return string(blackfriday.MarkdownCommon([]byte(v)))
	case []byte:
		return string(blackfriday.MarkdownCommon(v))
	}
	return ""
}

func jsonMenu(d interface{}) (string, error) {
	switch v := d.(type) {
	case string:
		return parse.JSONMenu(v)
	case []byte:
		return parse.JSONMenu(string(v))
	}

	return "", fmt.Errorf("jsonMenu requires string or []byte")

}

//Power Templates Takes a bunch a glob for a collection of templates, and then loads them all, adding the bonus functions to the templates abilities. Logs and Panics if templates don't parse.
func NewPowerTemplate(glob string, root string) *PowerTemplate {
	//Todo assign Sharer elsewhere

	t := template.New("")
	fMap := template.FuncMap{
		"tDict":     tDict,
		"randRange": RandRange,
		"md":        mdParse,
		"jsonMenu":  jsonMenu,
		"bSelect":   boolSelect,
	}

	blobMap, killer := blob.SafeBlobFuncs(root)
	for k, v := range blobMap {
		fMap[k] = v
	}
	t = t.Funcs(fMap)
	t, err := t.ParseGlob(glob)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return &PowerTemplate{t, killer}

}

/*
   This is a lazy Execution method. This will write the io.Writer with the execution of the template. It handles any error, by both writing it to the User, and also to std out.
*/
func Exec(t *template.Template, w io.Writer, tName string, data interface{}) {
	err := t.ExecuteTemplate(w, tName, data)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
}
