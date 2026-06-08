package main

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

type TSSParty struct {
	Id        int                       `json:"id"`
	PartyID   *tss.PartyID              `json:"party_id"`
	PreParams keygen.LocalPreParams     `json:"pre_params"`
	KeyShare  keygen.LocalPartySaveData `json:"key_share"`
}

func NewTSSParty(id int) (*TSSParty, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}

	partyId := tss.NewPartyID(fmt.Sprintf("party-%d", id), fmt.Sprintf("Party-%d", id), big.NewInt(int64(id)))

	return &TSSParty{
		Id:      id,
		PartyID: partyId,
	}, nil
}

func (t TSSParty) Print() {
	fmt.Printf("ID: %d\n", t.Id)
	fmt.Printf("PartyID: %+v\n", *t.PartyID)
	fmt.Printf("PreParams: %+v\n", t.PreParams)
	fmt.Printf("KeyShare: %+v\n", t.KeyShare)
}
