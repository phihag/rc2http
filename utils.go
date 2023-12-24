package main

import (
	"fmt"
	"os"
)

func FatalError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
