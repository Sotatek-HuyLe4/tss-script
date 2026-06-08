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
	fmt.Printf("TSS instance created successfully\n")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")

	/// STEP 2: Create TSS parties
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("STEP 2: Create TSS parties")
	err = tss.CreateParties()
	if err != nil {
		fmt.Printf("Failed to create TSS parties: %v\n", err)
		return
	}
	fmt.Printf("TSS parties created successfully\n")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")

	/// STEP 3: Generate Key
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("STEP 3: Generate Key")
	tss.GenerateKey()
	fmt.Printf("Key generated successfully\n")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")

	/// STEP 4: Sign Message
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("STEP 4: Sign Message")
	signatureData, err := tss.SignMessage("Hello, world!")
	if err != nil {
		fmt.Printf("Failed to sign message: %v\n", err)
		return
	}
	fmt.Printf("Message signed successfully: %+v\n", signatureData)
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")
}
