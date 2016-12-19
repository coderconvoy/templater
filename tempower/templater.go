package tempower

import (
	"errors"
	"fmt"
	"github.com/coderconvoy/templater/blob"
	"github.com/coderconvoy/templater/parse"
	"github.com/russross/blackfriday"
	"io"
	"math/rand"
	"path/filepath"
	"reflect"
	"text/template"
)

type PowerTemplate struct {
	*template.Template
	killer func()
}

func (pt PowerTemplate) Kill() {
	pt.killer()
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

func getN(n int, d interface{}) (interface{}, error) {
	//TODO consider adding support for maps

	if reflect.TypeOf(d).Kind() != reflect.Slice {
		return nil, fmt.Errorf("Not a slice")
	}

	s := reflect.ValueOf(d)
	if n < 0 {
		n = s.Len()
	}
	res := reflect.MakeSlice(reflect.TypeOf(d), 0, 0)
	l := s.Len()
	p := rand.Perm(l)
	for i := 0; i < n; i++ {
		res = reflect.Append(res, s.Index(p[i%l]))

	}

	return res.Interface(), nil

}

//Power Templates Takes a bunch a glob for a collection of templates, and then loads them all, adding the bonus functions to the templates abilities. Logs and Panics if templates don't parse.
func NewPowerTemplate(glob string, root string) (*PowerTemplate, error) {
	//Todo assign Sharer elsewhere

	t := template.New("")
	fMap := template.FuncMap{
		"tDict":     tDict,
		"randRange": RandRange,
		"md":        mdParse,
		"jsonMenu":  jsonMenu,
		"bSelect":   boolSelect,
		"getN":      getN,
	}

	blobMap, killer := blob.SafeBlobFuncs(root)
	for k, v := range blobMap {
		fMap[k] = v
	}
	t = t.Funcs(fMap)

	globArr, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	ar2 := make([]string, 0, 0)

	for _, v := range globArr {
		_, f := filepath.Split(v)
		if len(f) > 0 {
			if f[0] != '.' {
				ar2 = append(ar2, v)
			}
		}
	}
	t, err = t.ParseFiles(ar2...)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &PowerTemplate{t, killer}, nil

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
