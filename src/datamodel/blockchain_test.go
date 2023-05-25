package datamodel

import (
	"fmt"
	"math/big"
	"pedersen-commitment-transfer/src/pedersen"
	"testing"

	"github.com/bwesterb/go-ristretto"
)

func TestCreateAccount(t *testing.T) {
	// Set up dependencies
	var b Blockchain
	var vX ristretto.Scalar
	b.Init()
	fmt.Println("h is ", &b.h)
	fmt.Println("")
	fmt.Println("bindingFactor is ", &b.bindingFactor)
	// Define test case
	balance := int64(1000)
	expectedCommittedBalance := pedersen.CommitTo(&b.h, &b.bindingFactor, vX.SetBigInt(big.NewInt(balance)))
	fmt.Println("hhhhhheeeeeeiiii")
	// Call the function being tested
	account := b.CreateAccount(balance)

	// Verify the output
	if account.committedBalance != expectedCommittedBalance {
		t.Errorf("Unexpected committedBalance value. Got: %v, Expected: %v", account.committedBalance, expectedCommittedBalance)
	}
}
