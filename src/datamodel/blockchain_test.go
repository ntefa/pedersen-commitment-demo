package datamodel

import (
	"math/big"
	"pedersen-commitment-transfer/src/pedersen"
	"testing"

	"github.com/bwesterb/go-ristretto"
)

var _TestIsValidEncryption = struct {
	name        string
	amount      int64
	wrongH      ristretto.Point
	errorString string
}{
	name:        "Invalid encryption",
	amount:      100,
	wrongH:      pedersen.GenerateH(),
	errorString: "encryption not valid",
}

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

func TestIsValidEncryption(t *testing.T) {
	var b Blockchain
	var vX ristretto.Scalar
	b.Init()
	testCase := _TestIsValidEncryption

	wrongCommittedAmount := pedersen.CommitTo(&testCase.wrongH, &b.BindingFactor, vX.SetBigInt(big.NewInt(testCase.amount)))

	// Call the function being tested
	err := b.isValidEncryption(testCase.amount, wrongCommittedAmount)

	// Verify the output
	if err.Error() != testCase.errorString {
		t.Fatalf("Expected %v, but got: %v", testCase.errorString, err.Error())
	}
}
