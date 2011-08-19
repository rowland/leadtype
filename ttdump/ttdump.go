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
	if err == nil {
		fmt.Print(font)
	} else {
		fmt.Println(err)
	}
}
