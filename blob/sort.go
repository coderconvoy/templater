package blob

import "sort"

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
