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
	tX := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(big.NewInt(balance)))
	var account Account
	account.CommittedBalance = tX

	return &account
}

func (b *Blockchain) SetAccount(address string, account *Account) error {
	b.AddressList[address] = account
	return nil
}

//TODO decide whether transactionAmount should be encrypted or not
func (b *Blockchain) EncryptedTransaction(transactionAmount int64, senderAddress string) ristretto.Point {
	var vX ristretto.Scalar
	senderAccount := b.AddressList[senderAddress]
	//recipientAccount := b.addressList[recipientAddress] //TODO: add committed value to recipient
	amount := big.NewInt(transactionAmount)

	committedAmount := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(amount))
	fmt.Println("balance before sub", senderAccount.CommittedBalance)
	senderAccount.CommittedBalance.Sub(&senderAccount.CommittedBalance, &committedAmount)
	fmt.Println("balance after sub", senderAccount.CommittedBalance)
	return senderAccount.CommittedBalance
	//add money to recipient account
	//H := generateH()

}

func (b *Blockchain) isValidEncryption(x int64, eX ristretto.Point) error {
	var vX ristretto.Scalar
	value := big.NewInt(x)

	committedValue := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(value))

	if eX != committedValue {
		return fmt.Errorf("encryption not valid")
	}
	return nil
}

func (b *Blockchain) GetCommittedBalance(address string) ristretto.Point {
	account := b.AddressList[address]
	return account.CommittedBalance
}
