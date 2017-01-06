package tempower

import (
	"errors"
	"fmt"
	"github.com/coderconvoy/templater/blob"
	"github.com/coderconvoy/templater/parse"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"math/rand"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

type PowerTemplate struct {
	*template.Template
	root   string
	killer func()
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

	tMap := fileGetter(root)
	for k, v := range tMap {
		fMap[k] = v
	}

	tMap, killer := blob.SafeBlobFuncs(root)
	for k, v := range tMap {
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
	return &PowerTemplate{t, root, killer}, nil

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

//Select n random non repeating elements from slice d  returns error on d not slice
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

//Safelyfinds file from local scope
func fileGetter(root string) template.FuncMap {
	getFile := func(fname string) ([]byte, error) {
		p := path.Join(root, fname)
		if !strings.HasPrefix(p, root) {
			return []byte{}, fmt.Errorf("No upward pathing")
		}
		return ioutil.ReadFile(p)
	}

	mdFile := func(fname string) (string, error) {
		f, err := getFile(fname)
		if err != nil {
			return "", err
		}
		return mdParse(f), nil
	}

	getHeadedFile := func(fname string) (map[string]string, error) {
		f, err := getFile(fname)
		if err != nil {
			return map[string]string{}, err
		}
		return parse.Headed(f), nil
	}

	getHeadedMDFile := func(fname string) (map[string]string, error) {
		m, err := getHeadedFile(fname)
		if err != nil {
			return m, err
		}
		c, ok := m["contents"]
		if !ok {
			return m, fmt.Errorf("No Contents")
		}
		m["md"] = mdParse(c)
		return m, nil
	}

	return template.FuncMap{
		"File":         getFile,
		"mdFile":       mdFile,
		"headedFile":   getHeadedFile,
		"headedMDFile": getHeadedMDFile,
	}
}
