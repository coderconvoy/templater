//This file provides allows log to be a simple call in whatever form we expect.
package cfm



func Log(s string) {
}

func Logf(s string, d ...interface{}) {

	l := fmt.Sprintf(s, d...)
	Log(l)
}

func Logq(d ...interface){
	l := dbase.SLog(d...)
	Log(l)
}

func LogTo(l,s string){
}

func LogTof(l,s string){
}
