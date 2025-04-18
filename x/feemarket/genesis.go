package feemarket

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feemarketkeeper "github.com/EscanBE/everlast/x/feemarket/keeper"
	feemarkettypes "github.com/EscanBE/everlast/x/feemarket/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(
	ctx sdk.Context,
	k feemarketkeeper.Keeper,
	data feemarkettypes.GenesisState,
) []abci.ValidatorUpdate {
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the fee market module
func ExportGenesis(ctx sdk.Context, k feemarketkeeper.Keeper) *feemarkettypes.GenesisState {
	return &feemarkettypes.GenesisState{
		Params: k.GetParams(ctx),
	}
}
