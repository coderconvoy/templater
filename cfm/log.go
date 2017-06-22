//This file provides allows log to be a simple call in whatever form we expect.
package cfm

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/coderconvoy/dbase"
)

type logger interface {
	Log(s string)
	LogTo(l, s string)
}

//single, the core logger, will be used by all package log methods, to allow interchangeable Loggin
var single logger = FmtLogger{}

func SetLogger(l logger) {
	single = l

}

type FPathGetter interface {
	RootPath(...string) string
	GetFilePath(string, string) (string, error)
}

type FmtLogger struct{}

func (FmtLogger) Log(s string) {
	fmt.Println(s)
}

func (FmtLogger) LogTo(l, s string) {
	fmt.Println("Logto:", l, ":", s)
}

//LogData is for sending data through the channel in File Logger
type logdata struct {
	l, s string
}

//FileLogger : only make one, then keep it alive to do all logging
type FileLogger struct {
	getter FPathGetter
	ch     chan logdata
}

func NewFileLogger(fpg FPathGetter) FileLogger {
	ch := make(chan logdata, 20)
	go func() {
		for a := range ch {

			now := time.Now()
			fname := now.Format("060102")
			p := path.Join(a.l, fname+".log")
			err := os.MkdirAll(a.l, 0777)
			if err != nil {
				fmt.Println("Could not make dir:", p, err)
				continue
			}

			f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				fmt.Println("message not logged : ", err, "::", a.l)
				continue
			}

			line := now.Format("15:04:05") + "::" + a.s + "\n"
			_, err = f.WriteString(line)
			if err != nil {
				fmt.Println("message not logged: ", err, "::", a.l, a.s)
			}
			f.Close()
		}
	}()
	return FileLogger{
		fpg, ch,
	}
}

func (fl FileLogger) Log(s string) {
	loc := fl.getter.RootPath("logs/")
	fl.ch <- logdata{loc, s}
}

func (fl FileLogger) LogTo(l, s string) {
	fl.Log(l + "::" + s)
	loc, err := fl.getter.GetFilePath(l, "logs/")
	if err != nil {
		return
	}

	fl.ch <- logdata{loc, s}
}

// ----  Public Package methods  -----

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
