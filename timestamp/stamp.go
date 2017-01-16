// +build linux

package timestamp

import (
	"os"
	"syscall"
	"time"
)

func GetMod(fname string) (time.Time, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return time.Time{}, err
	}

	mTime := fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t)
	cTime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	aTime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))

	if mTime.After(cTime) && mTime.After(aTime) {
		return mTime, nil
	}
	if aTime.After(cTime) {
		return aTime, nil
	}
	return cTime, nil

}
