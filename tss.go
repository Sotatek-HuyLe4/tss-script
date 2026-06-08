package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
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

	// sort parties
	t.sortParties()

	// generate preparams
	t.generatePreParams()

	return nil
}

func (t *TSS) GenerateKey() {
	sortedPartyIds := make(tss.SortedPartyIDs, 0, t.TotalParties)
	for _, party := range t.Parties {
		sortedPartyIds = append(sortedPartyIds, party.PartyID)
	}

	outChs := make([]chan tss.Message, t.TotalParties)
	endChs := make([]chan *keygen.LocalPartySaveData, t.TotalParties)
	localParties := make([]*keygen.LocalParty, t.TotalParties)

	// init local parties
	for i, party := range t.Parties {
		outChs[i] = make(chan tss.Message, t.TotalParties*t.TotalParties)
		endChs[i] = make(chan *keygen.LocalPartySaveData, 1)

		ctx := tss.NewPeerContext(sortedPartyIds)
		params := tss.NewParameters(tss.S256(), ctx, party.PartyID, t.TotalParties, t.Threshold)
		localParties[i] = keygen.NewLocalParty(params, outChs[i], endChs[i], party.PreParams).(*keygen.LocalParty)
	}

	// start local parties
	for _, localParty := range localParties {
		go localParty.Start()
	}

	// route messages through all local parties
	var wg sync.WaitGroup
	for i := 0; i < t.TotalParties; i++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			for msg := range outChs[idx] {
				wireBytes, routing, _ := msg.WireBytes()

				if routing.IsBroadcast {
					// broadcast to all parties
					for j, localParty := range localParties {
						if j != idx {
							localParty.UpdateFromBytes(wireBytes, sortedPartyIds[idx], true)
						}
					}
				} else {
					// route to specific party
					for _, dest := range routing.To {
						for j, localParty := range localParties {
							if sortedPartyIds[j].Id == dest.Id {
								localParty.UpdateFromBytes(wireBytes, sortedPartyIds[idx], false)
							}
						}
					}
				}
			}
		}(i)
	}

	// collect result and save key share
	for i, party := range t.Parties {
		party.KeyShare = *<-endChs[i]
		close(outChs[i])
	}
}

func (t *TSS) sortParties() {
	unsortedPartyIds := make(tss.UnSortedPartyIDs, 0, t.TotalParties)
	for _, party := range t.Parties {
		unsortedPartyIds = append(unsortedPartyIds, party.PartyID)
	}

	tss.SortPartyIDs(unsortedPartyIds)
}

func (t *TSS) generatePreParams() {
	var wg sync.WaitGroup

	for _, party := range t.Parties {
		wg.Add(1)

		go func(party *TSSParty) {
			defer wg.Done()

			p, err := keygen.GeneratePreParams(2 * time.Minute)
			if err != nil {
				panic(err)
			}

			party.PreParams = *p
		}(party)
	}

	wg.Wait()
}
