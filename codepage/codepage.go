package codepage

type Range struct {
	firstCode  int
	lastCode   int
	entryCount int
	delta      int
}

type Ranges []Range

func (list Ranges) CharForCodepoint(cp int) (ch int, found bool) {
	low, high := 0, len(list)-1
	for low <= high {
		i := (low + high) / 2
		r := &list[i]
		if cp < r.firstCode {
			high = i - 1
			continue
		}
		if cp > r.lastCode {
			low = i + 1
			continue
		}
		return cp + r.delta, true
	}
	return 0, false
}
