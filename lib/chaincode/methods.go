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

func IsValidEncryption(ctx contractapi.TransactionContextInterface, x int64, committedAmount *ristretto.Point) error {

	H, bindingFactor, _, err := GetPedersenParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch pedersen encryption parameters: %v", err)
	}

	isValid := pedersen.Validate(x, *committedAmount, *H, *bindingFactor)
	fmt.Println(isValid)
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
	BindingFactorJSON, err := ctx.GetStub().GetState(PEDERSEN_BINDING_ID)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	}

	ZeroPedersenJSON, err := ctx.GetStub().GetState(PEDERSEN_ZERO_ID)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	}

	var H, zeroPedersen ristretto.Point
	var bindingFactor ristretto.Scalar

	//This case should not happen, param is passed in input to Init
	if HJSON == nil {
		H = ristretto.Point{}
		err = nil
	} else {
		err = H.UnmarshalBinary(HJSON)
		if err != nil {
			return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to unmarshal H : %v", err)
		}
	}

	//This case should not happen, param is passed in input to Init
	if BindingFactorJSON == nil {
		bindingFactor = ristretto.Scalar{}
		err = nil
	} else {
		err = bindingFactor.UnmarshalBinary(BindingFactorJSON)
		if err != nil {
			return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to unmarshal the Binding Factor : %v", err)
		}
	}

	//This case should not happen, param is passed in input to Init
	if ZeroPedersenJSON == nil {
		zeroPedersen = ristretto.Point{}
		err = nil
	} else {
		err = zeroPedersen.UnmarshalBinary(ZeroPedersenJSON)
		if err != nil {
			return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to unmarshal the vale for the committed zero : %v", err)
		}
	}

	return &H, &bindingFactor, &zeroPedersen, nil
}
