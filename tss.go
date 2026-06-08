package main

import (
	"errors"
	"fmt"
)

type TSS struct {
	TotalParties int         `json:"total_parties"`
	Threshold    int         `json:"threshold"`
	Parties      []*TSSParty `json:"parties"`
}

func NewTSS(totalParties, threshold int) (*TSS, error) {
	if totalParties == 0 || threshold == 0 || totalParties < threshold {
		return nil, errors.New("invalid total parties or threshold")
	}

	return &TSS{
		TotalParties: totalParties,
		Threshold:    threshold,
		Parties:      make([]*TSSParty, totalParties),
	}, nil
}

func (t TSS) Print() {
	fmt.Printf("Total Parties: %d\n", t.TotalParties)
	fmt.Printf("Threshold: %d\n", t.Threshold)
	for _, party := range t.Parties {
		if party != nil {
			fmt.Printf("Party %d:\n", party.Id)
			party.Print()
			fmt.Printf("\n")
		}
	}
}

func (t *TSS) CreateParties() error {
	for i := 0; i < t.TotalParties; i++ {
		party, err := NewTSSParty(i + 1)
		if err != nil {
			return err
		}

		t.Parties[i] = party
	}

	return nil
}
