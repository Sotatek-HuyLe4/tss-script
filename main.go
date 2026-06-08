package main

import (
	"fmt"
	"strings"
)

const (
	TOTAL_PARTIES = 3
	THRESHOLD     = 1
)

func main() {
	/// STEP 1: Create TSS instance
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("STEP 1: Create TSS instance")
	tss, err := NewTSS(TOTAL_PARTIES, THRESHOLD)
	if err != nil {
		fmt.Printf("Failed to create TSS instance: %v\n", err)
		return
	}
	fmt.Printf("TSS instance created successfully: %+v\n", tss)
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")

}
