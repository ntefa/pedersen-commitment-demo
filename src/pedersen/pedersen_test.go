package pedersen

import (
	"math/big"
	"testing"

	"github.com/bwesterb/go-ristretto"
	"github.com/stretchr/testify/assert"
)

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
		rX.Rand()
		H1 := GenerateH() // Secondary point on the Curve
		var H2 ristretto.Point

		amount2 := big.NewInt(testcase.amount2)

		// Transfer amount of 5 tokens
		amount2Committed := CommitTo(&H1, &rX, vX.SetBigInt(amount2)) //5 encrypted tokens
		// Alice 10 - 5 = 5
		rY.Rand()
		amount1 := big.NewInt(testcase.amount1)
		amount1Committed := CommitTo(&H1, &rY, vY.SetBigInt(amount1))
		assert.NotEqual(t, amount1Committed, amount2Committed, "Should not be equal")
		var sumCommitted ristretto.Point
		sumCommitted.Add(&amount1Committed, &amount2Committed)

		if testcase.isError {
			H2 = testcase.H
			checksumCommitted := AddPrivately(&H2, &rX, &rY, amount1, amount2)
			assert.False(t, checksumCommitted.Equals(&sumCommitted), "Should not be equal")
		} else {
			H2 = H1
			checksumCommitted := AddPrivately(&H2, &rX, &rY, amount1, amount2)
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

		amount2 := big.NewInt(testcase.amount2)

		// Transfer amount of 5 tokens
		amount2Committed := CommitTo(&H1, &rX, vX.SetBigInt(amount2)) //5 encrypted tokens
		// Alice 10 - 5 = 5
		rY.Rand()
		amount1 := big.NewInt(testcase.amount1)
		amount1Committed := CommitTo(&H1, &rY, vY.SetBigInt(amount1))
		assert.NotEqual(t, amount1Committed, amount2Committed, "Should not be equal")
		var sumCommitted ristretto.Point
		sumCommitted.Sub(&amount1Committed, &amount2Committed)

		if testcase.isError {
			H2 = testcase.H
			checksumCommitted := SubPrivately(&H2, &rX, &rY, amount1, amount2)
			assert.False(t, checksumCommitted.Equals(&sumCommitted), "Should not be equal")
		} else {
			H2 = H1
			checksumCommitted := SubPrivately(&H2, &rX, &rY, amount1, amount2)
			assert.True(t, checksumCommitted.Equals(&sumCommitted), "Should be equal")
		}
	}
}
