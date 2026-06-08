package main

import (
	"encoding/hex"
	"errors"
	"math/big"
)

func StringToBigInt(message string) (*big.Int, error) {
	// convert string to hex
	hexString := hex.EncodeToString([]byte(message))

	// convert hex string to big int
	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(hexString, 16)
	if !ok {
		return nil, errors.New("failed to convert hex string to big int")
	}

	return bigInt, nil
}
