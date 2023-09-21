package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type TxInformation struct {
	Sender string
	Amount []byte
}

func createTxInfo(sender string, amount ristretto.Point) (TxInformation, error) {
	amountBytes, err := amount.MarshalBinary()
	if err != nil {
		return TxInformation{}, err
	}
	txInfo := TxInformation{
		Sender: sender,
		Amount: amountBytes,
	}
	return txInfo, nil
}

func storeTxInfo(stub shim.ChaincodeStubInterface, txId string, sender string, amount ristretto.Point) error {
	txInfo, err := createTxInfo(sender, amount)
	if err != nil {
		return err
	}

	transferDetailsBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal: %v", err)
	}

	err = stub.PutState(txId, transferDetailsBytes)
	if err != nil {
		return err
	}

	return nil
}

func getTxInfo(stub shim.ChaincodeStubInterface, TxId string) (string, ristretto.Point, error) {
	TxInfoBytes, err := stub.GetState(TxId)
	if err != nil {
		return "", ristretto.Point{}, fmt.Errorf("failed to read transaction information from world state: %v", err)
	}
	var txInfo TxInformation
	json.Unmarshal(TxInfoBytes, &txInfo)
	if txInfo.Amount == nil {
		return "", ristretto.Point{}, fmt.Errorf("temporary account has no balance")
	}
	var committedAmount ristretto.Point                  //variable to store the current committed balance of sender
	err = committedAmount.UnmarshalBinary(txInfo.Amount) //recipient should be clientId
	if err != nil {
		return "", ristretto.Point{}, fmt.Errorf("error unmarshalling")
	}
	return txInfo.Sender, committedAmount, nil
}

// func (TxInformation) Marshal() ([]byte, error) {

// }

// func (TxInformation) Marshal([]byte, error) {
// 	//we need to convert amount to bytes first due to behavious of ristretto points
// 	amountBytes, err := TxInformation.Amount.MarshalBinary()
// 	if err != nil {
// 		t.Fatalf("Error is: %v", err)
// 	}

// 	testStruct := TestStruct2{
// 		Teststring:     "blabla",
// 		Testcommitment: committedAmountBytes,
// 	}
// 	result, err := json.Marshal(testStruct)
// 	if err != nil {
// 		t.Fatalf("Error is: %v", err)
// 	}

// }
