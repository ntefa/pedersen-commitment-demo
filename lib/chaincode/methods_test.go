package chaincode

import (
	"fmt"
	"math/big"
	"pedersen-commitment-transfer/lib/tests/testsfakes"
	"pedersen-commitment-transfer/src/pedersen"
	"testing"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/assert"
)

var _TestEncryption = []struct {
	name        string
	amount      int64
	wrongAmount int64
	isError     bool
	errorString string
}{
	{
		name:        "OK",
		amount:      100,
		wrongAmount: 100,
		isError:     false,
	},
	{
		name:        "Invalid encryption",
		amount:      100,
		wrongAmount: 50,
		isError:     true,
		errorString: "encryption not valid",
	},
}

func TestInitPedersen(t *testing.T) {
	ctx := &testsfakes.FakeTestTransactionContextInterface{}
	stub := &testsfakes.FakeTestChaincodeStubInterface{}
	ctx.GetStubStub = func() shim.ChaincodeStubInterface {
		return stub
	}
	stub.GetTxIDStub = func() string {
		return "TxidTest"
	}
	H, bindingFactor, _ := generateRandomCommitment(100)

	InitPedersen(ctx, H, bindingFactor)

	//Let's check that putstate was called correctly.
	_, HJSON := stub.PutStateArgsForCall(0)
	var H_Test ristretto.Point
	H_Test.UnmarshalBinary(HJSON)

	if !H_Test.Equals(&H) {
		t.Fatal("Error")
	}

	_, BindingFactorJSON := stub.PutStateArgsForCall(1)
	var bindingFactor_Test ristretto.Scalar
	bindingFactor_Test.UnmarshalBinary(BindingFactorJSON)

	if !bindingFactor_Test.Equals(&bindingFactor) {
		t.Fatal("Error")
	}
}

func TestGetPedersenParams(t *testing.T) {
	ctx := &testsfakes.FakeTestTransactionContextInterface{}
	stub := &testsfakes.FakeTestChaincodeStubInterface{}
	ctx.GetStubStub = func() shim.ChaincodeStubInterface {
		return stub
	}
	stub.GetTxIDStub = func() string {
		return "TxidTest"
	}
	H, bindingFactor, zeroPedersen := generateRandomCommitment(100)
	HJSON, _ := H.MarshalBinary()
	stub.PutState(PEDERSEN_H_ID, HJSON)
	BindingFactorJSON, _ := bindingFactor.MarshalBinary()
	stub.PutState(PEDERSEN_BINDING_ID, BindingFactorJSON)
	ZeroPedersenJSON, _ := zeroPedersen.MarshalBinary()
	stub.GetStateReturnsOnCall(0, HJSON, nil)
	stub.GetStateReturnsOnCall(1, BindingFactorJSON, nil)
	stub.GetStateReturnsOnCall(2, ZeroPedersenJSON, nil)

	var H2, zeroPedersen2 *ristretto.Point
	var bindingFactor2 *ristretto.Scalar
	H2, bindingFactor2, zeroPedersen2, _ = GetPedersenParams(ctx)

	if !H2.Equals(&H) {
		t.Fatal("Error")
	}

	if !zeroPedersen2.Equals(&zeroPedersen) {
		t.Fatal("Error")
	}

	if !bindingFactor2.Equals(&bindingFactor) {
		t.Fatal("Error")
	}

	var htest ristretto.Point
	htest.UnmarshalBinary(nil)
	fmt.Println(htest)

}

func TestIsValidEncryption(t *testing.T) {
	ctx := &testsfakes.FakeTestTransactionContextInterface{}
	stub := &testsfakes.FakeTestChaincodeStubInterface{}
	ctx.GetStubStub = func() shim.ChaincodeStubInterface {
		return stub
	}
	stub.GetTxIDStub = func() string {
		return "TxidTest"
	}
	for i, testcase := range _TestEncryption {
		t.Run(testcase.name, func(t *testing.T) {

			//It uses GetPedersenParams under the hood, hence similar mocking methods.
			H, bindingFactor, committedAmount := generateRandomCommitment(testcase.amount)
			HJSON, _ := H.MarshalBinary()
			stub.PutState(PEDERSEN_H_ID, HJSON)
			BindingFactorJSON, _ := bindingFactor.MarshalBinary()
			stub.PutState(PEDERSEN_BINDING_ID, BindingFactorJSON)
			committedAmountJSON, _ := committedAmount.MarshalBinary()

			//We need one set of calls per iteration
			stub.GetStateReturnsOnCall(i, HJSON, nil)
			stub.GetStateReturnsOnCall(i+1, BindingFactorJSON, nil)
			stub.GetStateReturnsOnCall(i+2, committedAmountJSON, nil)
			//***********************************************************************

			err := IsValidEncryption(ctx, testcase.wrongAmount, &committedAmount)
			if !testcase.isError {
				if err != nil {
					t.Fatalf("Error is: %v", err)
				}
			} else {
				assert.EqualError(t, err, testcase.errorString)
			}
		})
	}

}

func generateRandomCommitment(amount int64) (ristretto.Point, ristretto.Scalar, ristretto.Point) {

	var rX, vX ristretto.Scalar
	H1 := pedersen.GenerateH() // Secondary point on the Curve
	rX.Rand()
	amountBig := big.NewInt(amount)
	amountCommitted := pedersen.CommitTo(&H1, &rX, vX.SetBigInt(amountBig))
	return H1, rX, amountCommitted
}
