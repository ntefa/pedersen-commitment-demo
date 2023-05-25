package test

import (
	"pedersen-commitment-transfer/src/datamodel"
	"testing"
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

func TestTransfer(t *testing.T) {

	for _, testCase := range _TestTransfer {
		var b datamodel.Blockchain

		b.Init()
		account := b.CreateAccount(testCase.initialBalance)
		b.SetAccount(testCase.address, account)
		//b = createBlockchain(b, "aaa")
	}
}

// func createBlockchain(b *datamodel.Blockchain, address string, account *datamodel.Account) *datamodel.Blockchain {
// 	b.Init()
// 	b.CreateAccount(amount)
// 	b.SetAccount(address, account)
// 	return b
// }
