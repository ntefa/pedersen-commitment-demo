package datamodel

import (
	"fmt"
	"math/big"

	"pedersen-commitment-transfer/src/pedersen"

	"github.com/bwesterb/go-ristretto"
)

func (b *Blockchain) Init() {
	b.setH()

	b.setBindingFactor()

	b.AddressList = map[string]*Account{}
}
func (b *Blockchain) CreateAccount(balance int64) *Account {
	var vX ristretto.Scalar
	//committed amount
	tX := pedersen.CommitTo(&b.h, &b.bindingFactor, vX.SetBigInt(big.NewInt(balance)))
	var account Account
	account.committedBalance = tX

	return &account
}

func (b *Blockchain) SetAccount(address string, account *Account) error {
	b.AddressList[address] = account
	return nil
}

func (b *Blockchain) encryptedTransaction(transactionAmount int64, eX ristretto.Point, Y uint, senderAddress string, recipientAddress string) {
	var vX ristretto.Scalar
	senderAccount := b.AddressList[senderAddress]
	//recipientAccount := b.addressList[recipientAddress] //TODO: add committed value to recipient
	amount := big.NewInt(transactionAmount)

	committedAmount := pedersen.CommitTo(&b.h, &b.bindingFactor, vX.SetBigInt(amount))
	senderAccount.committedBalance.Sub(&senderAccount.committedBalance, &committedAmount)
	//H := generateH()

}

func (b *Blockchain) isValidEncryption(x int64, eX ristretto.Point) error {
	var vX ristretto.Scalar
	value := big.NewInt(x)

	committedValue := pedersen.CommitTo(&b.h, &b.bindingFactor, vX.SetBigInt(value))

	if eX != committedValue {
		return fmt.Errorf("encryption not valid")
	}
	return nil
}
