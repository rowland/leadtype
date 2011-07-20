package main

import (
	"../../pdf/_obj/go-pdf/pdf"
	"os"
	"exec"
)

const name = "test_001_empty_doc.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter(f)
	doc.Close()
	f.Close()
	exec.Command("open", name).Start()
}
