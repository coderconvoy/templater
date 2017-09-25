package cfm

import (
	"errors"
	"path"
	"strings"
)

func SafeJoin(a ...string) (string, error) {
	if len(a) == 0 {
		return "", nil
	}

	res := path.Join(a...)
	if !strings.HasPrefix(res, a[0]) {
		return a[0], errors.New("No Upward pathing")
	}
	return res, nil
}
