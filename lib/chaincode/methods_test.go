package chaincode

//TODO: add test - look how to mock context etc.

// mockchaincode := new(chaincodetest.MockChainCode)
// mockStub := shimtest.NewMockStub("TestJoinService Stub", mockchaincode)
// mockStub.MockTransactionStart("TestTXN_1")
// ctx := trxcontext.GetNewCtx(mockStub)
// controller := new(Controller)
// controller.Ctx = ctx
// mockStub.ChannelID = "TestChannel"

import (
	"math/big"
	"pedersen-commitment-transfer/lib/tests/testsfakes"
	"pedersen-commitment-transfer/src/pedersen"
	"testing"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

var _TestPedersen = struct {
	name        string
	amount      int64
	wrongH      ristretto.Point
	errorString string
}{
	name:        "Invalid encryption",
	amount:      100,
	wrongH:      pedersen.GenerateH(),
	errorString: "encryption not valid",
}

func TestInitPedersen(t *testing.T) {
	sc := SmartContract{}
	ctx := &testsfakes.FakeTestTransactionContextInterface{}
	stub := &testsfakes.FakeTestChaincodeStubInterface{}
	//stub := NewMockStub
	ctx.GetStubStub = func() shim.ChaincodeStubInterface {
		return stub
	}
	stub.GetTxIDStub = func() string {
		return "TxidTest"
	}
	H, bindingFactor, _ := generateRandomCommitment(100)

	sc.InitPedersen(ctx, H, bindingFactor) //test that it is calling put state with those input params

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

func generateRandomCommitment(amount int64) (ristretto.Point, ristretto.Scalar, ristretto.Point) {

	var rX, vX ristretto.Scalar
	H1 := pedersen.GenerateH() // Secondary point on the Curve
	rX.Rand()
	amountBig := big.NewInt(amount)
	amountCommitted := pedersen.CommitTo(&H1, &rX, vX.SetBigInt(amountBig))
	return H1, rX, amountCommitted
}
