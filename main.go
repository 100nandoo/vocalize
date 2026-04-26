package main

import (
	"fmt"
	"os"

	"github.com/100nandoo/inti/cmd"
)

func main() {
	cmd.WebFS = webFS
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
