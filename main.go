package main

import (
	"fmt"
	"os"
)

func main() {
	err := Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
