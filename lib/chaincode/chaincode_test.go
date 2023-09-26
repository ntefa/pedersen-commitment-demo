package chaincode

import (
	"encoding/json"
	"pedersen-commitment-transfer/lib/tests/testsfakes"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

var _TestInit = []struct {
	name        string
	data        string
	isError     bool
	errorString string
}{
	{
		name:    "OK",
		data:    "Initialized",
		isError: false,
	},
	{
		name:        "Not initialized",
		data:        "",
		isError:     true,
		errorString: "encryption not valid",
	},
}

func TestCheckInit(t *testing.T) {
	ctx := &testsfakes.FakeTestTransactionContextInterface{}
	stub := &testsfakes.FakeTestChaincodeStubInterface{}
	ctx.GetStubStub = func() shim.ChaincodeStubInterface {
		return stub
	}
	stub.GetTxIDStub = func() string {
		return "TxidTest"
	}

	for _, testcase := range _TestInit {
		t.Run(testcase.name, func(t *testing.T) {
			if !testcase.isError {
				dataJSON, _ := json.Marshal(testcase.data)
				stub.GetStateReturns(dataJSON, nil)
				resp, _ := checkInitialized(ctx)

				if resp != true {
					t.Fatalf("response should be true")
				}
			} else {
				stub.GetStateReturns(nil, nil)
				resp, _ := checkInitialized(ctx)

				if resp != false {
					t.Fatal("response should be false")
				}
			}

		})
	}
}
