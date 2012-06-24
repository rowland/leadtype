// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"leadtype/ttf"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ttdump fontname [table1 [table2...]]")
		os.Exit(-1)
	}
	fontname := os.Args[1]
	fmt.Println("ttdump", fontname)
	font, err := ttf.LoadFont(fontname)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if len(os.Args) < 3 {
		font.Dump(os.Stdout, "all")
	} else {
		for i := 2; i < len(os.Args); i++ {
			font.Dump(os.Stdout, os.Args[i])
		}
	}
}
