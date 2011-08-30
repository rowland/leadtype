package main

import (
	"fmt"
	"os"
	"../ttf/_obj/go-pdf/ttf"
)

func main() {
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
