package test

import (
	"math/big"
	"pedersen-commitment-transfer/src/datamodel"
	"pedersen-commitment-transfer/src/pedersen"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _TestTransfer = []struct {
	senderAddress           string
	recipientAddress        string
	senderInitialBalance    int64
	recipientInitialBalance int64
	amountSent              int64
}{
	{
		senderAddress:           "0x111",
		recipientAddress:        "0x999",
		senderInitialBalance:    10,
		recipientInitialBalance: 20,
		amountSent:              5,
	},
}

func TestCommittedTransaction(t *testing.T) {

	//Setup H and binding factor (should they be fixed values in blockchain?)
	var b datamodel.Blockchain
	b.Init()

	for _, testCase := range _TestTransfer {

		//Create an account with an initial committed balance
		senderAccount := b.CreateAccount(testCase.senderInitialBalance)
		recipientAccount := b.CreateAccount(testCase.recipientInitialBalance)

		//Assign to the specific address the account
		b.SetAccount(testCase.senderAddress, senderAccount)
		b.SetAccount(testCase.recipientAddress, recipientAccount)

		senderInitialBalanceCommitted := b.AddressList[testCase.senderAddress].GetBalance()
		recipientInitialBalanceCommitted := b.AddressList[testCase.recipientAddress].GetBalance()

		//Trigger transcacton. Subtract committed account from initial committed balance
		senderCommittedAmount_afterTransaction, recipientCommittedAmount_afterTransaction := b.EncryptedTransaction(testCase.amountSent, testCase.senderAddress, testCase.recipientAddress)

		//Assert change of state in the balance
		assert.NotEqual(t, senderInitialBalanceCommitted, senderCommittedAmount_afterTransaction, "Should not be equal")
		assert.NotEqual(t, recipientInitialBalanceCommitted, recipientCommittedAmount_afterTransaction, "Should not be equal")

		amountSent := big.NewInt(testCase.amountSent)
		senderInitialBalance := big.NewInt(testCase.senderInitialBalance)
		recipientInitialBalance := big.NewInt(testCase.recipientInitialBalance)

		checkAmountCommitted := pedersen.SubPrivately(&b.H, &b.BindingFactor, &b.BindingFactor, senderInitialBalance, amountSent)
		assert.True(t, checkAmountCommitted.Equals(&senderCommittedAmount_afterTransaction), "Should be equal")
		checkAmountCommitted = pedersen.AddPrivately(&b.H, &b.BindingFactor, &b.BindingFactor, recipientInitialBalance, amountSent)
		assert.True(t, checkAmountCommitted.Equals(&recipientCommittedAmount_afterTransaction), "Should be equal")
	}
}
