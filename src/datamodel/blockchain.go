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
	account.committedBalance = tX

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
	fmt.Println("balance before sub", senderAccount.committedBalance)
	senderAccount.committedBalance.Sub(&senderAccount.committedBalance, &committedAmount)
	fmt.Println("balance after sub", senderAccount.committedBalance)
	return senderAccount.committedBalance
	//add money to recipient account
	//H := generateH()

}

func (b *Blockchain) isValidEncryption(x int64, committedAmount ristretto.Point) error {
	var vX ristretto.Scalar
	value := big.NewInt(x)

	committedValue := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(value))

	if committedAmount != committedValue {
		return fmt.Errorf("encryption not valid")
	}
	return nil
}

func (b *Blockchain) GetCommittedBalanceByAddress(address string) ristretto.Point {
	account := b.AddressList[address]
	return account.GetBalance()
}

func (account *Account) GetBalance() ristretto.Point {
	return account.committedBalance
}
