package evm

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	evmtypes "github.com/EscanBE/evermint/v12/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// EthSigVerificationDecorator validates an ethereum signatures
type EthSigVerificationDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthSigVerificationDecorator creates a new EthSigVerificationDecorator
func NewEthSigVerificationDecorator(ek EVMKeeper) EthSigVerificationDecorator {
	return EthSigVerificationDecorator{
		evmKeeper: ek,
	}
}

// AnteHandle validates checks that the registered chain id is the same as the one on the message, and
// that the signer address matches the one defined on the message.
// It's not skipped for RecheckTx, because it set `From` address which is critical from other ante handler to work.
// Failure in RecheckTx will prevent tx to be included into block, especially when CheckTx succeed, in which case user
// won't see the error message.
func (esvd EthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	chainID := esvd.evmKeeper.ChainID()
	evmParams := esvd.evmKeeper.GetParams(ctx)
	chainCfg := evmParams.GetChainConfig()
	ethCfg := chainCfg.EthereumConfig(chainID)
	blockNum := big.NewInt(ctx.BlockHeight())
	signer := ethtypes.MakeSigner(ethCfg, blockNum)

	{
		msgEthTx := tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)

		ethTx := msgEthTx.AsTransaction()
		if !ethTx.Protected() {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrNotSupported,
				"rejected unprotected Ethereum transaction. Please EIP155 sign your transaction to protect it against replay-attacks")
		}

		sender, err := signer.Sender(ethTx)
		if err != nil {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrorInvalidSigner,
				"couldn't retrieve sender address from the ethereum transaction: %s",
				err.Error(),
			)
		}

		senderBech32 := sdk.AccAddress(sender.Bytes()).String()
		if msgEthTx.From != senderBech32 {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrorInvalidSigner,
				"mis-match sender address %s vs %s (%s) from signer",
				msgEthTx.From, senderBech32, sender.Hex(),
			)
		}
	}

	return next(ctx, tx, simulate)
}
