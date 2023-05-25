package utils

import (
	"fmt"

	"github.com/bwesterb/go-ristretto"
)

type Sender struct {
	Balance    uint
	privateKey ristretto.Scalar
}

func (s *Sender) SetBalance(balance uint) {
	s.Balance = balance
}

func (s *Sender) triggerTransfer(to string, amount uint) error {
	if amount > s.Balance {
		return fmt.Errorf(" You don't have enough funds to transfer")
	}
	s.Balance -= amount

	return nil
}

// func (b *Blockchain) updateRecipient(y int, address string) error {
// 	recipient := b.addressList[address]
// 	eY := recipient.committedBalance
// }
