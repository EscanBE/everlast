package types

import "encoding/hex"

const (
	GasVerifyEIP712 = 200_000
)

var PseudoCodePrecompiled []byte

func init() {
	var err error
	PseudoCodePrecompiled, err = hex.DecodeString(
		// ABI string of "Custom Precompiled Contract"
		"0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001b437573746f6d20507265636f6d70696c656420436f6e74726163740000000000",
	)
	if err != nil {
		panic(err)
	}
}
