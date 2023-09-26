package chaincode

import (
	"encoding/json"
	"fmt"
	"math/big"
	"pedersen-commitment-transfer/src/pedersen"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const PEDERSEN_ID = "PEDERSEN"
const PEDERSEN_H_ID = "PEDERSEN_H"
const PEDERSEN_BINDING_ID = "PEDERSEN_BINDING"
const PEDERSEN_ZERO_ID = "PEDERSEN_ZERO"

func IsValidEncryption(ctx contractapi.TransactionContextInterface, x int64, committedAmount *ristretto.Point) error {

	//Fetch pedersen parameters from state
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

	pedersenVariables := createPedersenVariables(HJSON, BindingFactorJSON, ZeroPedersenJSON)
	pedersenVariablesJSON, err := json.Marshal(pedersenVariables)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(PEDERSEN_ID, pedersenVariablesJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// ReadAsset returns the asset stored in the world state with given id.
func GetPedersenParams(ctx contractapi.TransactionContextInterface) (*ristretto.Point, *ristretto.Scalar, *ristretto.Point, error) {
	pedersenVariablesJson, err := ctx.GetStub().GetState(PEDERSEN_ID)
	if err != nil {
		return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	var pedersenVariables PedersenVariables
	err = json.Unmarshal(pedersenVariablesJson, &pedersenVariables)
	if err != nil {
		fmt.Errorf("failed to unmarshal")
	}

	// HJSON, err := ctx.GetStub().GetState(PEDERSEN_H_ID)
	// if err != nil {
	// 	return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	// }
	// BindingFactorJSON, err := ctx.GetStub().GetState(PEDERSEN_BINDING_ID)
	// if err != nil {
	// 	return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	// }

	// ZeroPedersenJSON, err := ctx.GetStub().GetState(PEDERSEN_ZERO_ID)
	// if err != nil {
	// 	return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to read from world state: %v", err)
	// }

	var H, zeroPedersen ristretto.Point
	var bindingFactor ristretto.Scalar

	//This case should not happen, param is passed in input to Init
	if pedersenVariables.H_bytes == nil {
		H = ristretto.Point{}
		err = nil
	} else {
		err = H.UnmarshalBinary(pedersenVariables.H_bytes)
		if err != nil {
			return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to unmarshal H : %v", err)
		}
	}

	//This case should not happen, param is passed in input to Init
	if pedersenVariables.BindingFactor_bytes == nil {
		bindingFactor = ristretto.Scalar{}
		err = nil
	} else {
		err = bindingFactor.UnmarshalBinary(pedersenVariables.BindingFactor_bytes)
		if err != nil {
			return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to unmarshal the Binding Factor : %v", err)
		}
	}

	//This case should not happen, param is passed in input to Init
	if pedersenVariables.ZeroCommitted_bytes == nil {
		zeroPedersen = ristretto.Point{}
		err = nil
	} else {
		err = zeroPedersen.UnmarshalBinary(pedersenVariables.ZeroCommitted_bytes)
		if err != nil {
			return &ristretto.Point{}, &ristretto.Scalar{}, &ristretto.Point{}, fmt.Errorf("failed to unmarshal the vale for the committed zero : %v", err)
		}
	}

	return &H, &bindingFactor, &zeroPedersen, nil
}
