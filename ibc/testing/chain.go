package ibctesting

import (
	"testing"

	"github.com/EscanBE/everlast/constants"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/cosmos/ibc-go/v8/testing/mock"

	"github.com/EscanBE/everlast/crypto/ethsecp256k1"
	evertypes "github.com/EscanBE/everlast/types"
)

// ChainIDPrefix defines the default chain ID prefix for our test chains
var ChainIDPrefix = constants.TestnetChainID

func init() {
	ibctesting.ChainIDPrefix = ChainIDPrefix
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on
// counterparty chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the Coordinator in order to ensure synchrony between chains.
// Each update of any chain increments the block header time for all chains by 5 seconds.
func NewTestChain(t *testing.T, coord *ibctesting.Coordinator, chainID string) *ibctesting.TestChain {
	// generate validator private/public key
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})
	signers := make(map[string]cmttypes.PrivValidator)
	signers[pubKey.Address().String()] = privVal

	// generate genesis account
	senderPrivKey, err := ethsecp256k1.GenerateKey()
	if err != nil {
		panic(err)
	}

	baseAcc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)

	amount := sdk.TokensFromConsensusPower(1, evertypes.PowerReduction)

	balance := banktypes.Balance{
		Address: baseAcc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(constants.BaseDenom, amount)),
	}

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{baseAcc}, chainID, balance)

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := app.GetTxConfig()

	// create an account to send transactions from
	chain := &ibctesting.TestChain{
		TB:            t,
		Coordinator:   coord,
		ChainID:       chainID,
		App:           app,
		CurrentHeader: header,
		QueryServer:   app.GetIBCKeeper(),
		TxConfig:      txConfig,
		Codec:         app.AppCodec(),
		Vals:          valSet,
		Signers:       signers,
		SenderPrivKey: senderPrivKey,
		SenderAccount: baseAcc,
		NextVals:      valSet,
	}

	coord.CommitBlock(chain)

	return chain
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *Path {
	path := NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
}
