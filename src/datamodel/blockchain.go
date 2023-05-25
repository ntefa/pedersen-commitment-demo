package blockchain

import (
	"fmt"
	"math/big"

	"github.com/bwesterb/go-ristretto"
	"github.com/threehook/go-pedersen-commitment/src/pedersen"
)

func (b *Blockchain) initBlockchain() {
	b.setH()
	b.setBindingFactor()
}

func (b *Blockchain) setAccount(address string, account Account) error {
	b.addressList[address] = account
	return nil
}
func (b *Blockchain) encryptedTransaction(transactionAmount int64, eX ristretto.Point, Y uint, senderAddress string, recipientAddress string) {
	var vX ristretto.Scalar
	senderAccount := b.addressList[senderAddress]
	//recipientAccount := b.addressList[recipientAddress] //TODO: add committed value to recipient
	amount := big.NewInt(transactionAmount)

	committedAmount := pedersen.CommitTo(&b.H, &b.bindingFactor, vX.SetBigInt(amount))
	senderAccount.committedBalance.Sub(&senderAccount.committedBalance, &committedAmount)
	//H := generateH()

}

func (b *Blockchain) isValidEncryption(x int64, eX ristretto.Point) error {
	var vX ristretto.Scalar
	value := big.NewInt(x)

	committedValue := pedersen.CommitTo(&b.H, &b.bindingFactor, vX.SetBigInt(value))

	if eX != committedValue {
		return fmt.Errorf("encryption not valid")
	}
	return nil
}
