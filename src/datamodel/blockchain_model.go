package datamodel

import (
	"pedersen-commitment-transfer/src/pedersen"

	"github.com/bwesterb/go-ristretto"
)

type Blockchain struct {
	AddressList   map[string]*Account
	h             ristretto.Point
	bindingFactor ristretto.Scalar
}

type Account struct {
	committedBalance ristretto.Point
}

// H is a constant value for the blockchain. It needs to be setup. If H varies the state of the blockchain is not consistent.
func (b *Blockchain) setH() {
	b.h = pedersen.GenerateH()
}

func (b *Blockchain) setBindingFactor() {
	b.bindingFactor.Rand()
}
