package chaincode

import (
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

	H, bindingFactor, _, err := GetPedersenParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch pedersen encryption parameters: %v", err)
	}

	isValid := pedersen.Validate(x, *committedAmount, *H, *bindingFactor)

	if !isValid {
		return fmt.Errorf("encryption not valid")
	}
	return nil
}

// InitLedger adds a base set of assets to the ledger
func InitPedersen(ctx contractapi.TransactionContextInterface, H ristretto.Point, bindingFactor ristretto.Scalar) error {

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

// ReadAsset returns the asset stored in the world state with given id.
func GetPedersenParams(ctx contractapi.TransactionContextInterface) (*ristretto.Point, *ristretto.Scalar, *ristretto.Point, error) {

	HJSON, err := ctx.GetStub().GetState(PEDERSEN_H_ID)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	if HJSON == nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("the asset %s does not exist", PEDERSEN_H_ID)
	}

	BindingFactorJSON, err := ctx.GetStub().GetState(PEDERSEN_BINDING_ID)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	if BindingFactorJSON == nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("the asset %s does not exist", PEDERSEN_H_ID)
	}

	ZeroPedersenJSON, err := ctx.GetStub().GetState(PEDERSEN_ZERO_ID)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	if ZeroPedersenJSON == nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("the asset %s does not exist", PEDERSEN_H_ID)
	}

	var H, zeroPedersen ristretto.Point
	var bindingFactor ristretto.Scalar

	err = H.UnmarshalBinary(HJSON)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("Failed to unmarshal asset")
	}

	err = bindingFactor.UnmarshalBinary(BindingFactorJSON)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("Failed to unmarshal asset")
	}

	err = zeroPedersen.UnmarshalBinary(ZeroPedersenJSON)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("Failed to unmarshal asset")
	}

	return &H, &bindingFactor, &zeroPedersen, nil
}
