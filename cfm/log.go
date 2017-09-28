//This file provides allows log to be a simple call in whatever form we expect.
package cfm

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/coderconvoy/dbase"
)

//single, the core logger, will be used by all package log methods, to allow interchangeable Loggin
var single logger = logger{}

// ----  Public Package methods  -----

func SetLogger(man *Manager, af string) {
	single = logger{man: man, allFolder: af}
}

func Log(s string) {
	single.Log(s)
}

func Logf(s string, d ...interface{}) {
	l := fmt.Sprintf(s, d...)
	single.Log(l)
}

func Logq(d ...interface{}) {
	l := dbase.SLog(d...)
	single.Log(l)
}

func LogTo(l, s string) {
	single.LogTo(l, s)
}

func LogTof(l, s string, d ...interface{}) {
	f := fmt.Sprintf(s, d...)
	single.LogTo(l, f)
}
func LogToq(l string, d ...interface{}) {
	f := dbase.SLog(d...)
	single.LogTo(l, f)
}

// --- Under the hood ----
//by using locks within goroutines, we protect from deadlock, and allow the function to return quickly, without waiting for the file write.

type logger struct {
	man       *Manager
	allFolder string
	sync.Mutex
}

//Log uses a go routine with a mutex for filewrites
func (lg logger) Log(s string) {
	go func() {
		lg.Lock()
		defer lg.Unlock()

		if lg.allFolder == "" {
			fmt.Println("Logger not Set", s)
			return
		}
		err := logToFolder(lg.allFolder, s)
		if err != nil {
			fmt.Println("Logging err", err, s)
		}
	}()
}

//LogTo uses go-routine with a mutex for fileWrites
func (lg logger) LogTo(host, s string) {
	ps := host + "::" + s
	lg.Log(ps)
	go func() {
		lg.Lock()
		defer lg.Unlock()
		if lg.man == nil {
			fmt.Println("Manager not set: ", s)
			return
		}
		cf, err := lg.man.GetConfig(host)
		if err != nil {
			lg.Log("Logging Host not found: " + s)
			return
		}
		err = logToFolder(path.Join(cf.Folder, "logs"), s)
		if err != nil {
			lg.Log("Could not access host log folder," + host + "," + cf.Folder + "," + s)
			return
		}
	}()
}

func logToFolder(folder string, s string) error {
	now := time.Now()
	fname := now.Format("060102")
	p := path.Join(folder, fname)
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	line := now.Format("15:04:05") + "::" + s + "\n"
	_, err = f.WriteString(line)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}
