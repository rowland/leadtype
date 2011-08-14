package main

import (
	"fmt"
	"os"
	"../ttf/_obj/go-pdf/ttf"
)

func main() {
	fontname := os.Args[1]
	fmt.Println("ttdump", fontname)
	font, _ := ttf.LoadFont(fontname)
	fmt.Print(font)
}
