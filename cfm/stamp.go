package timestamp

import (
	"os"
	"time"
)

func GetModified(fname string) (time.Time, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}
