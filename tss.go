package main

import "errors"

const (
	TOTAL_PARTIES = 3
	THRESHOLD     = 2
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
