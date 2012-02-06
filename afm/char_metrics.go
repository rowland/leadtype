package afm

type CharMetric struct {
	code  rune
	width int32
	name  string
}

type CharMetrics []CharMetric

func NewCharMetrics(count int) CharMetrics {
	return make(CharMetrics, count)
}

func (cms CharMetrics) Len() int {
	return len(cms)
}

// Less returns whether the element with index i should sort
// before the element with index j.
func (cms CharMetrics) Less(i, j int) bool {
	return cms[i].code < cms[j].code
}

// Swap swaps the elements with indexes i and j.
func (cms CharMetrics) Swap(i, j int) {
	cms[i], cms[j] = cms[j], cms[i]
}

func (cms CharMetrics) ForRune(codepoint rune) *CharMetric {
	low, high := 0, len(cms)-1
	for low <= high {
		i := (low + high) / 2
		if codepoint > cms[i].code {
			low = i + 1
			continue
		}
		if codepoint < cms[i].code {
			high = i - 1
			continue
		}
		return &cms[i]
	}
	return nil
}
