package chaincode

import (
	"encoding/json"
	"fmt"
	"math/big"
	"pedersen-commitment-transfer/src/pedersen"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const PEDERSEN_H_ID = "PEDERSEN_H"
const PEDERSEN_BINDING_ID = "PEDERSEN_BINDING"
const PEDERSEN_ZERO_ID = "PEDERSEN_ZERO"

// TODO: better to store as single params, each with its key. Look at totalsupply get in token-erc-20
type PedersenParams struct {
	ID            string           `json:"ID"`
	H             ristretto.Point  `json:"H"`
	BindingFactor ristretto.Scalar `json:"BindingFactor"`
	ZeroCommitted ristretto.Point  `json:"ZeroCommitted"` //TODO: think if the value of zero committed should be public, stored in the ledger or some other way
	test          int
}

func IsValidEncryption(ctx contractapi.TransactionContextInterface, x int64, committedAmount *ristretto.Point) error {

	pedersenParams, err := GetPedersenParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch pedersen encryption parameters: %v", err)
	}

	H := pedersenParams.H
	bindingFactor := pedersenParams.BindingFactor
	isValid := pedersen.Validate(x, *committedAmount, H, bindingFactor)

	if !isValid {
		return fmt.Errorf("encryption not valid")
	}
	return nil
}

// InitLedger adds a base set of assets to the ledger
func (SmartContract) InitPedersen(ctx contractapi.TransactionContextInterface, H ristretto.Point, bindingFactor ristretto.Scalar) error {

	var vX ristretto.Scalar

	zero := big.NewInt(0)

	zeroCommitted := pedersen.CommitTo(&H, &bindingFactor, vX.SetBigInt(zero))

	HJSON, err := H.MarshalBinary()
	if err != nil {
		return err
	}

	BindingFactorJSON, err := bindingFactor.MarshalBinary()
	if err != nil {
		return err
	}

	ZeroPedersenJSON, err := zeroCommitted.MarshalBinary()
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(PEDERSEN_H_ID, HJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	err = ctx.GetStub().PutState(PEDERSEN_BINDING_ID, BindingFactorJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	err = ctx.GetStub().PutState(PEDERSEN_ZERO_ID, ZeroPedersenJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// // ReadAsset returns the asset stored in the world state with given id.
func GetPedersenParams(ctx contractapi.TransactionContextInterface) (*PedersenParams, error) {
	assetJSON, err := ctx.GetStub().GetState(PEDERSEN_H_ID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", PEDERSEN_H_ID)
	}

	var asset PedersenParams
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}
