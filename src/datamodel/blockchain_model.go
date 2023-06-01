package datamodel

import (
	"pedersen-commitment-transfer/src/pedersen"

	"github.com/bwesterb/go-ristretto"
)

type Blockchain struct {
	AddressList   map[string]*Account
	H             ristretto.Point
	BindingFactor ristretto.Scalar
}

type Account struct {
	CommittedBalance ristretto.Point
}

// H is a constant value for the blockchain. It needs to be setup. If H varies the state of the blockchain is not consistent.
func (b *Blockchain) setH() {
	b.H = pedersen.GenerateH()
}

func (b *Blockchain) setBindingFactor() {
	b.BindingFactor.Rand()
}
