package main

import (
	noloopclosure "github.com/fatanugraha/noloopclosure"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(noloopclosure.Analyzer)
}
