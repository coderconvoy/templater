//This file provides allows log to be a simple call in whatever form we expect.
package cfm

import (
	"fmt"

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
	RootPath(string) string
	GetFilePath(string, string) string
}

type FmtLogger struct{}

func (FmtLogger) Log(s string) {
	fmt.Println(s)
}

func (FmtLogger) LogTo(l, s string) {
	fmt.Println("Logto:", l, ":", s)
}

type FileLogger struct{ getter FPathGetter }

func (fl FileLogger) Log(s string) {
	loc := fl.getter.RootPath("logs/")
	fl.LogIn(loc, s)
}

func (fl FileLogger) LogTo(l, s string) {
	loc := fl.getter.GetFilePath(l, "logs/")
	fl.LogIn(loc, s)
}

func (fl FileLogger) LogIn(l, s string) {
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
