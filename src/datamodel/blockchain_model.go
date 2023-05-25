package blockchain

import (
	"github.com/bwesterb/go-ristretto"
	"github.com/threehook/go-pedersen-commitment/src/pedersen"
)

type Blockchain struct {
	addressList   map[string]Account
	H             ristretto.Point
	bindingFactor ristretto.Scalar
}

type Account struct {
	committedBalance ristretto.Point
}

// H is a constant value for the blockchain. It needs to be setup. If H varies the state of the blockchain is not consistent.
func (b *Blockchain) setH() {
	b.H = pedersen.GenerateH()
}

func (b *Blockchain) setBindingFactor() {
	b.bindingFactor.Rand()
}
