package templater

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"text/template"
)

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

func GetSharedFileText(libname string) string {
	return string(GetSharedFile(libname))
}

func RandRange(l, h int) int {
	return rand.Intn(h-l) + l
}

/*
   Takes a bunch a glob for a collection of templates, and then loads them all, adding the bonus functions to the templates abilities. Logs and Panics if templates don't parse.
*/
func PowerTemplates(glob string) *template.Template {
	t := template.New("")
	fMap := template.FuncMap{
		"tDict":          tDict,
		"sharedFileText": GetSharedFileText,
		"sharedMD":       GetSharedMD,
		"htmlMenu":       HTMLMenu,
		"jsonMenu":       JSONMenu,
		"randRange":      RandRange,
		"getDirList":     GetDirList,
		"getHeadedMD":    GetSharedHeadedMD,
	}
	t = t.Funcs(fMap)
	t, err := t.ParseGlob(glob)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return t
}

/*
   This is the core Execution method. This will write the io.Writer with the execution of the template. It handles any error, by both writing it to the User, and also to std out.
*/
func Exec(t *template.Template, w io.Writer, tName string, data interface{}) {
	err := t.ExecuteTemplate(w, tName, data)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
}
