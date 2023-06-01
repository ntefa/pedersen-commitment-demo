package datamodel

import (
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

	// Define test case
	balance := int64(1000)
	expectedCommittedBalance := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(big.NewInt(balance)))
	// Call the function being tested
	account := b.CreateAccount(balance)

	// Verify the output
	if account.CommittedBalance != expectedCommittedBalance {
		t.Errorf("Unexpected committedBalance value. Got: %v, Expected: %v", account.CommittedBalance, expectedCommittedBalance)
	}
}
