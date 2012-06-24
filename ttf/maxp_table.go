// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"fmt"
	"io"
	"os"
)

type maxpTable struct {
	version               Fixed
	numGlyphs             uint16
	maxPoints             uint16
	maxContours           uint16
	maxComponentPoints    uint16
	maxComponentContours  uint16
	maxZones              uint16
	maxTwilightPoints     uint16
	maxStorage            uint16
	maxFunctionDefs       uint16
	maxInstructionDefs    uint16
	maxStackElements      uint16
	maxSizeOfInstructions uint16
	maxComponentElements  uint16
	maxComponentDepth     uint16
}

func (table *maxpTable) init(file io.ReadSeeker, entry *tableDirEntry) (err error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	// No advantage to using a buffered reader here.
	if err = table.version.Read(file); err != nil {
		return
	}
	err = readValues(file,
		&table.numGlyphs,
		&table.maxPoints,
		&table.maxContours,
		&table.maxComponentPoints,
		&table.maxComponentContours,
		&table.maxZones,
		&table.maxTwilightPoints,
		&table.maxStorage,
		&table.maxFunctionDefs,
		&table.maxInstructionDefs,
		&table.maxStackElements,
		&table.maxSizeOfInstructions,
		&table.maxComponentElements,
		&table.maxComponentDepth,
	)
	return
}

func (table *maxpTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "maxp Table")
	fmt.Fprintf(wr, "version               = %g\n", table.version.Tof64())
	fmt.Fprintf(wr, "numGlyphs             = %d\n", table.numGlyphs)
	fmt.Fprintf(wr, "maxPoints             = %d\n", table.maxPoints)
	fmt.Fprintf(wr, "maxContours           = %d\n", table.maxContours)
	fmt.Fprintf(wr, "maxComponentPoints    = %d\n", table.maxComponentPoints)
	fmt.Fprintf(wr, "maxComponentContours  = %d\n", table.maxComponentContours)
	fmt.Fprintf(wr, "maxZones              = %d\n", table.maxZones)
	fmt.Fprintf(wr, "maxTwilightPoints     = %d\n", table.maxTwilightPoints)
	fmt.Fprintf(wr, "maxStorage            = %d\n", table.maxStorage)
	fmt.Fprintf(wr, "maxFunctionDefs       = %d\n", table.maxFunctionDefs)
	fmt.Fprintf(wr, "maxInstructionDefs    = %d\n", table.maxInstructionDefs)
	fmt.Fprintf(wr, "maxStackElements      = %d\n", table.maxStackElements)
	fmt.Fprintf(wr, "maxSizeOfInstructions = %d\n", table.maxSizeOfInstructions)
	fmt.Fprintf(wr, "maxComponentElements  = %d\n", table.maxComponentElements)
	fmt.Fprintf(wr, "maxComponentDepth     = %d\n", table.maxComponentDepth)
}
