package blob

import (
	"sort"
	"strings"
)

type LessF func(PageInfo, PageInfo) bool

type SortyBlob struct {
	contents []PageInfo
	lessf    LessF
}

func (sb SortyBlob) Len() int {
	return len(sb.contents)
}

func (sb SortyBlob) Less(a, b int) bool {
	return sb.lessf(sb.contents[a], sb.contents[b])
}

func (sb SortyBlob) Swap(a, b int) {
	sb.contents[a], sb.contents[b] = sb.contents[b], sb.contents[a]
}

func Sort(dt []PageInfo, f LessF) {
	b := SortyBlob{
		dt, f,
	}
	sort.Sort(b)
}

func ByName(ascend bool) LessF {
	return func(a, b PageInfo) bool {
		if ascend {
			return a.FName < b.FName
		}
		return b.FName < a.FName
	}
}

func ByDate(ascend bool) LessF {
	return func(a, b PageInfo) bool {
		if ascend {
			return a.Date.Before(b.Date)
		}
		return b.Date.Before(a.Date)
	}
}

func ByProp(pname string, ascend bool) LessF {
	return func(a, b PageInfo) bool {
		//empty all goes to one end
		adat, ok := a.extra[pname]
		if !ok {
			return ascend
		}
		bdat, ok := b.extra[pname]
		if !ok {
			return ascend
		}
		//standard responses
		if ascend {
			return adat < bdat
		}
		return bdat < adat
	}
}

func BlobSort(dt []PageInfo, sMode string) {
	if sMode == "" {
		sMode = "-date"
	}
	ascend := true
	if strings.HasPrefix(sMode, "-") {
		ascend = false
		sMode = strings.TrimPrefix(sMode, "-")
	}

	switch sMode {
	case "name":
		Sort(dt, ByName(ascend))
	case "date":
		Sort(dt, ByDate(ascend))
	default:
		Sort(dt, ByProp(sMode, ascend))
	}
}
