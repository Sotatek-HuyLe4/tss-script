package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/resharing"
	"github.com/bnb-chain/tss-lib/v2/tss"
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

func VerifySignature(message string, signatureData *common.SignatureData, keyShare keygen.LocalPartySaveData) (bool, error) {
	// convert message to big int
	msgBigInt, err := StringToBigInt(message)
	if err != nil {
		return false, err
	}
	msgDigest := msgBigInt.Bytes()

	// signature data
	signatureR := new(big.Int).SetBytes(signatureData.R)
	signatureS := new(big.Int).SetBytes(signatureData.S)

	// pubkey data
	pubkey := keyShare.ECDSAPub
	ecdsaPubkey := &ecdsa.PublicKey{
		Curve: tss.S256(),
		X:     pubkey.X(),
		Y:     pubkey.Y(),
	}

	// verify signature
	valid := ecdsa.Verify(ecdsaPubkey, msgDigest, signatureR, signatureS)
	if !valid {
		return false, errors.New("signature verification failed")
	}

	return true, nil

}

// routeResharingMessage mirrors tss-lib's ecdsa/resharing/local_party_test.go routing.
func routeResharingMessage(
	msg tss.Message,
	oldCommittee, newCommittee []*resharing.LocalParty,
	errCh chan<- *tss.Error,
) {
	dest := msg.GetTo()
	if dest == nil {
		panic("resharing message has nil destination")
	}

	updateParty := func(party tss.Party, msg tss.Message) {
		if party.PartyID().KeyInt().Cmp(msg.GetFrom().KeyInt()) == 0 {
			return
		}

		bz, _, err := msg.WireBytes()
		if err != nil {
			errCh <- party.WrapError(err)
			return
		}

		pMsg, err := tss.ParseWireMessage(bz, msg.GetFrom(), msg.IsBroadcast())
		if err != nil {
			errCh <- party.WrapError(err)
			return
		}
		if _, err := party.Update(pMsg); err != nil {
			errCh <- err
		}
	}

	if msg.IsToOldCommittee() || msg.IsToOldAndNewCommittees() {
		for _, destP := range dest[:len(oldCommittee)] {
			updateParty(oldCommittee[destP.Index], msg)
		}
	}
	
	if !msg.IsToOldCommittee() || msg.IsToOldAndNewCommittees() {
		for _, destP := range dest {
			updateParty(newCommittee[destP.Index], msg)
		}
	}
}
