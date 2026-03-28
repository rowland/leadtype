package ltml

type SplitResult struct {
	Head Widget
	Tail Widget
}

type Splittable interface {
	SplitForHeight(avail float64, w Writer) (*SplitResult, error)
}
