package types

import (
	"fmt"

	sdktxtypes "github.com/cosmos/cosmos-sdk/types/tx"

	abci "github.com/cometbft/cometbft/abci/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"

	evmtypes "github.com/EscanBE/everlast/x/evm/types"
	feemarkettypes "github.com/EscanBE/everlast/x/feemarket/types"
)

// QueryClient defines a gRPC Client used for:
//   - Transaction simulation
//   - EVM module queries
//   - Fee market module queries
type QueryClient struct {
	sdktxtypes.ServiceClient
	evmtypes.QueryClient
	FeeMarket feemarkettypes.QueryClient
}

// NewQueryClient creates a new gRPC query client
func NewQueryClient(clientCtx client.Context) *QueryClient {
	return &QueryClient{
		ServiceClient: sdktxtypes.NewServiceClient(clientCtx),
		QueryClient:   evmtypes.NewQueryClient(clientCtx),
		FeeMarket:     feemarkettypes.NewQueryClient(clientCtx),
	}
}

// GetProof performs an ABCI query with the given key and returns a merkle proof.
// The desired CometBFT height to perform the query should be set in the client context.
// The query will be performed at one below this height (at the IAVL version)
// in order to obtain the correct merkle proof.
// Proof queries at height less than or equal to 2 are not supported.
// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
func (QueryClient) GetProof(clientCtx client.Context, storeKey string, key []byte) ([]byte, *tmcrypto.ProofOps, error) {
	height := clientCtx.Height
	// ABCI queries at height less than or equal to 2 are not supported.
	// Base app does not support queries for height less than or equal to 1.
	// Therefore, a query at height 2 would be equivalent to a query at height 3
	if height <= 2 {
		return nil, nil, fmt.Errorf("proof queries at height <= 2 are not supported")
	}

	abciReq := abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", storeKey),
		Data:   key,
		Height: height,
		Prove:  true,
	}

	abciRes, err := clientCtx.QueryABCI(abciReq)
	if err != nil {
		return nil, nil, err
	}

	return abciRes.Value, abciRes.ProofOps, nil
}
