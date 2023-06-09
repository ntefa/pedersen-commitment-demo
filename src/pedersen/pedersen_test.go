package pedersen

import (
	"math/big"
	"testing"

	"github.com/bwesterb/go-ristretto"
	"github.com/stretchr/testify/assert"
)

//TODO: this implementation of pedersen uses signed integers, hence we can have a negative commitment after subtraction
//TODO: check if valid on wallet/client side, raise error there otherwise trigger smart contract method
var _TestCommittedValues = []struct {
	name    string
	H       ristretto.Point
	amount1 int64
	amount2 int64
	isError bool
}{
	{
		name:    "Ok",
		isError: false,
		amount1: 10,
		amount2: 5,
	},
	{
		name:    "Different H",
		H:       GenerateH(),
		amount1: 10,
		amount2: 5,
		isError: true,
	},
}

// Should commit to a sum of two values
func TestAddCommittedValues(t *testing.T) {

	for _, testcase := range _TestCommittedValues {

		var rX, rY, vX, vY ristretto.Scalar
		H1 := GenerateH() // Secondary point on the Curve

		// Commit amount1
		rX.Rand()
		amount1 := big.NewInt(testcase.amount1)
		amount1Committed := CommitTo(&H1, &rX, vX.SetBigInt(amount1))

		// Commit amount2
		rY.Rand()
		amount2 := big.NewInt(testcase.amount2)
		amount2Committed := CommitTo(&H1, &rY, vY.SetBigInt(amount2))

		//Check committed values are different
		assert.NotEqual(t, amount1Committed, amount2Committed, "Should not be equal")

		var H2 ristretto.Point

		//Compute committed sum
		var sumCommitted ristretto.Point
		sumCommitted.Add(&amount1Committed, &amount2Committed)

		if testcase.isError {
			H2 = testcase.H
			//Check that sum was correct
			checkSumCommitted := AddPrivately(&H2, &rY, &rX, amount1, amount2)
			assert.False(t, checkSumCommitted.Equals(&sumCommitted), "Should not be equal")
		} else {
			H2 = H1
			checksumCommitted := AddPrivately(&H2, &rY, &rX, amount1, amount2)
			assert.True(t, checksumCommitted.Equals(&sumCommitted), "Should be equal")
		}
	}
}

func TestSubCommittedValues(t *testing.T) {

	for _, testcase := range _TestCommittedValues {
		var rX, rY, vX, vY ristretto.Scalar
		rX.Rand()
		H1 := GenerateH() // Secondary point on the Curve
		var H2 ristretto.Point

		amount1 := big.NewInt(testcase.amount1)

		// Transfer amount of 5 tokens
		amount1Committed := CommitTo(&H1, &rX, vX.SetBigInt(amount1)) //5 encrypted tokens

		rY.Rand()
		amount2 := big.NewInt(testcase.amount2)
		amount2Committed := CommitTo(&H1, &rY, vY.SetBigInt(amount2))
		assert.NotEqual(t, amount1Committed, amount2Committed, "Should not be equal")

		var difCommitted ristretto.Point
		difCommitted.Sub(&amount1Committed, &amount2Committed)

		if testcase.isError {
			H2 = testcase.H
			checkdifCommitted := SubPrivately(&H2, &rY, &rX, amount1, amount2)
			assert.False(t, checkdifCommitted.Equals(&difCommitted), "Should not be equal")
		} else {
			H2 = H1
			checkdifCommitted := SubPrivately(&H2, &rY, &rX, amount1, amount2)
			assert.True(t, checkdifCommitted.Equals(&difCommitted), "Should be equal")
		}
	}
}

//Note that Add/SubPrivately are built so that the private keys in the signature are swapped out, i.e. rY,rX.
//Addition is invariant to that, but subtraction is affected.
