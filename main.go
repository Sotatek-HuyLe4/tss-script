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
	pubkey := tss.Parties[0].KeyShare.ECDSAPub
	fmt.Printf("	Public key X = %x\n", pubkey.X())
	fmt.Printf("	Public key Y = %x\n", pubkey.Y())
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
	fmt.Printf("Message signed successfully\n")
	fmt.Printf("	Signature.R: %+x\n", signatureData.R)
	fmt.Printf("	Signature.S: %+x\n", signatureData.S)
	// verify signature
	for _, party := range tss.Parties {
		isValid, err := VerifySignature("Hello, world!", signatureData, party.KeyShare)
		if err != nil {
			fmt.Printf("Failed to verify signature: %v\n", err)
			return
		}
		fmt.Printf("Signature verified successfully for party %d: %t\n", party.Id, isValid)
	}
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")

	/// STEP 5: Re-sharing Key
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("STEP 5: Re-sharing Key")
	err = tss.ReSharingKey(3, 4)
	if err != nil {
		fmt.Printf("Failed to re-share key: %v\n", err)
		return
	}
	fmt.Printf("Key re-shared successfully\n")
	pubkey = tss.Parties[0].KeyShare.ECDSAPub
	fmt.Printf("	Public key X = %x\n", pubkey.X())
	fmt.Printf("	Public key Y = %x\n", pubkey.Y())
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")

	/// STEP 6: Re-sign Message
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("STEP 6: Re-sign Message")
	signatureData, err = tss.SignMessage("Hello, world!")
	if err != nil {
		fmt.Printf("Failed to sign message: %v\n", err)
		return
	}
	fmt.Printf("Message signed successfully\n")
	fmt.Printf("	Signature.R: %+x\n", signatureData.R)
	fmt.Printf("	Signature.S: %+x\n", signatureData.S)
	// verify signature
	for _, party := range tss.Parties {
		isValid, err := VerifySignature("Hello, world!", signatureData, party.KeyShare)
		if err != nil {
			fmt.Printf("Failed to verify signature: %v\n", err)
			return
		}
		fmt.Printf("Signature verified successfully for party %d: %t\n", party.Id, isValid)
	}
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\n\n")
}
