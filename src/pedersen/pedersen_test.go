package pedersen

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/bwesterb/go-ristretto"
	"github.com/stretchr/testify/assert"
)

// Should commit to a sum of two values
func CommitToSuccess(t *testing.T) {

	var rX, rY, vX, vY ristretto.Scalar
	rX.Rand()
	H := GenerateH() // Secondary point on the Curve
	five := big.NewInt(5)

	// Transfer amount of 5 tokens
	tC := CommitTo(&H, &rX, vX.SetBigInt(five)) //5 encrypted tokens
	// Alice 10 - 5 = 5
	rY.Rand()
	fmt.Println("rY (second private key) is ", rY)
	ten := big.NewInt(10)
	aC1 := CommitTo(&H, &rY, vY.SetBigInt(ten))
	assert.NotEqual(t, aC1, tC, "Should not be equal")
	var aC2 ristretto.Point
	aC2.Sub(&aC1, &tC)

	checkAC2 := SubPrivately(&H, &rX, &rY, ten, five)
	assert.True(t, checkAC2.Equals(&aC2), "Should be equal")
}

// Should fail if not using the correct blinding factors
func CommitToFails(t *testing.T) {

	var rX, rY, vX, vY ristretto.Scalar
	rX.Rand()
	H := GenerateH() // Secondary point on the Curve
	five := big.NewInt(5)

	// Transfer amount of 5 tokens
	tC := CommitTo(&H, &rX, vX.SetBigInt(five))

	// Alice 10 - 5 = 5
	rY.Rand()
	ten := big.NewInt(10)
	aC1 := CommitTo(&H, &rY, vY.SetBigInt(ten))
	assert.NotEqual(t, aC1, tC, "They should not be equal")
	var aC2 ristretto.Point
	aC2.Sub(&aC1, &tC)

	// Create different (and wrong) binding factors
	rX.Rand()
	rY.Rand()
	checkAC2 := SubPrivately(&H, &rX, &rY, ten, five)
	assert.False(t, checkAC2.Equals(&aC2), "Should not be equal")
}
