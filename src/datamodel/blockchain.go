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
func (b *Blockchain) EncryptedTransaction(transactionAmount int64, senderAddress string, recipientAddress string) (ristretto.Point, ristretto.Point) {
	var vX ristretto.Scalar
	senderAccount := b.AddressList[senderAddress]
	recipientAccount := b.AddressList[recipientAddress] //TODO: add committed value to recipient
	amount := big.NewInt(transactionAmount)

	committedAmount := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(amount))
	//avoid double spending
	senderAccount.committedBalance.Sub(&senderAccount.committedBalance, &committedAmount)
	recipientAccount.committedBalance.Add(&recipientAccount.committedBalance, &committedAmount)

	return senderAccount.committedBalance, recipientAccount.committedBalance
	//add money to recipient account

}

func (b *Blockchain) isValidEncryption(x int64, committedAmount *ristretto.Point) error {

	isValid := pedersen.Validate(x, *committedAmount, b.H, b.BindingFactor)
	//committedValue := pedersen.CommitTo(&b.H, &b.BindingFactor, vX.SetBigInt(value))

	if !isValid {
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
