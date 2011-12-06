package codepage

type CharRange struct {
	firstCode  int
	lastCode   int
	entryCount int
	delta      int
}

type CharRanges []CharRange

func (list CharRanges) CharForCodepoint(cp int) (ch int, found bool) {
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

type CodepageRange struct {
	firstCode  int
	lastCode   int
	entryCount int
	codepage   string
}

type CodepageRanges []CodepageRange

func (list CodepageRanges) CodepageForCodepoint(cp int) (codepage string, found bool) {
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
		return r.codepage, true
	}
	return "", false
}
