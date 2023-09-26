package chaincode

import (
	"encoding/json"
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

var _TestRistrettoMarshalling = []struct {
	name    string
	isError bool
}{
	{
		name:    "OK",
		isError: false,
	},
	{
		name:    "Invalid encryption",
		isError: true,
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
	H, bindingFactor, _ := generateRandomCommitment(0)

	InitPedersen(ctx, H, bindingFactor)

	//Prepare expected values
	var vX ristretto.Scalar

	zero := big.NewInt(0)

	zeroCommitted := pedersen.CommitTo(&H, &bindingFactor, vX.SetBigInt(zero))

	//Let's check that putstate was called correctly.
	_, initPedersenPut := stub.PutStateArgsForCall(0)

	var pedersenVariables_got PedersenVariables
	json.Unmarshal(initPedersenPut, &pedersenVariables_got)

	var H_got ristretto.Point
	var bindingFactor_got ristretto.Scalar
	var zeroPedersen_got ristretto.Point

	H_got.UnmarshalBinary(pedersenVariables_got.H_bytes)
	bindingFactor_got.UnmarshalBinary(pedersenVariables_got.BindingFactor_bytes)
	zeroPedersen_got.UnmarshalBinary(pedersenVariables_got.ZeroCommitted_bytes)

	if !H_got.Equals(&H) {
		t.Fatal("Error")
	}

	if !zeroPedersen_got.Equals(&zeroCommitted) {
		t.Fatal("Error")
	}

	if !bindingFactor_got.Equals(&bindingFactor) {
		t.Fatalf("Error: expected %v, got %v", bindingFactor, bindingFactor_got)
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
	H, bindingFactor, zeroPedersen := generateRandomCommitment(0)
	HJSON, _ := H.MarshalBinary()
	BindingFactorJSON, _ := bindingFactor.MarshalBinary()
	ZeroPedersenJSON, _ := zeroPedersen.MarshalBinary()
	pedersenVariables := createPedersenVariables(HJSON, BindingFactorJSON, ZeroPedersenJSON)
	pedersenVariablesJson, _ := json.Marshal(pedersenVariables)
	stub.PutState(PEDERSEN_ID, pedersenVariablesJson)

	stub.GetStateReturnsOnCall(0, pedersenVariablesJson, nil)

	var H2, zeroPedersen2 *ristretto.Point
	var bindingFactor2 *ristretto.Scalar
	H2, bindingFactor2, zeroPedersen2, err := GetPedersenParams(ctx)

	if err != nil {
		t.Fatal(err)
	}
	if !H2.Equals(&H) {
		t.Fatal("Error")
	}

	if !zeroPedersen2.Equals(&zeroPedersen) {
		t.Fatal("Error")
	}

	if !bindingFactor2.Equals(&bindingFactor) {
		t.Fatalf("Error: expected %v, got %v", bindingFactor, bindingFactor2)
	}

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
			BindingFactorJSON, _ := bindingFactor.MarshalBinary()
			committedAmountJSON, _ := committedAmount.MarshalBinary()
			pedersenVariables := createPedersenVariables(HJSON, BindingFactorJSON, committedAmountJSON)
			pedersenVariablesJson, _ := json.Marshal(pedersenVariables)

			stub.GetStateReturnsOnCall(i, pedersenVariablesJson, nil)

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
