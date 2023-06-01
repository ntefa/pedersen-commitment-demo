package test

import (
	"math/big"
	"pedersen-commitment-transfer/src/datamodel"
	"pedersen-commitment-transfer/src/pedersen"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _TestTransfer = []struct {
	address        string
	initialBalance int64
	amountSent     int64
}{
	{
		address:        "0x111",
		initialBalance: 10,
		amountSent:     5,
	},
}

func TestCommittedTransaction(t *testing.T) {

	//Setup H and binding factor (should they be fixed values in blockchain?)
	var b datamodel.Blockchain
	b.Init()

	for _, testCase := range _TestTransfer {

		//Create an account with an initial committed balance
		account := b.CreateAccount(testCase.initialBalance)
		//Assign to the specific address the account
		b.SetAccount(testCase.address, account)
		initialBalanceCommitted := b.AddressList[testCase.address].CommittedBalance

		//Trigger transcacton. Subtract committed account from initial committed balance
		committedAmount_afterTransaction := b.EncryptedTransaction(testCase.amountSent, testCase.address)

		//Assert change of state in the balance
		assert.NotEqual(t, initialBalanceCommitted, committedAmount_afterTransaction, "Should not be equal")

		amountSent := big.NewInt(testCase.amountSent)
		initialBalance := big.NewInt(testCase.initialBalance)
		checkAmountCommitted := pedersen.SubPrivately(&b.H, &b.BindingFactor, &b.BindingFactor, initialBalance, amountSent)
		assert.True(t, checkAmountCommitted.Equals(&committedAmount_afterTransaction), "Should be equal")
	}
}
