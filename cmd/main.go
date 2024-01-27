package main

import (
	"os"

	"github.com/taylormonacelli/hisrabbit"
)

func main() {
	code := hisrabbit.Execute()
	os.Exit(code)
}
