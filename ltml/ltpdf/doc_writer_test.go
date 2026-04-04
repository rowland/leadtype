package ltpdf_test

import (
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/ltml/ltpdf"
)

var _ ltml.Writer = (*ltpdf.DocWriter)(nil)
