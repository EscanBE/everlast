package ledger_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cometbft/cometbft/version"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	chainapp "github.com/EscanBE/everlast/app"
	"github.com/EscanBE/everlast/app/helpers"
	"github.com/EscanBE/everlast/app/params"
	"github.com/EscanBE/everlast/constants"
	"github.com/EscanBE/everlast/crypto/hd"
	"github.com/EscanBE/everlast/tests/integration/ledger/mocks"
	utiltx "github.com/EscanBE/everlast/testutil/tx"
	"github.com/EscanBE/everlast/utils"

	clientkeys "github.com/EscanBE/everlast/client/keys"
	appkeyring "github.com/EscanBE/everlast/crypto/keyring"
	feemarkettypes "github.com/EscanBE/everlast/x/feemarket/types"
	cosmosledger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	sdk "github.com/cosmos/cosmos-sdk/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var s *LedgerTestSuite

type LedgerTestSuite struct {
	suite.Suite

	app *chainapp.EverLast
	ctx sdk.Context

	ledger       *mocks.SECP256K1
	accRetriever *mocks.AccountRetriever

	accAddr sdk.AccAddress

	privKey types.PrivKey
	pubKey  types.PubKey
}

func TestLedger(t *testing.T) {
	s = new(LedgerTestSuite)
	suite.Run(t, s)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Ledger Suite")
}

func (suite *LedgerTestSuite) SetupTest() {
	var (
		err     error
		ethAddr common.Address
	)

	suite.ledger = mocks.NewSECP256K1(s.T())

	ethAddr, s.privKey = utiltx.NewAddrKey()

	s.Require().NoError(err)
	suite.pubKey = s.privKey.PubKey()

	suite.accAddr = sdk.AccAddress(ethAddr.Bytes())
}

func (suite *LedgerTestSuite) SetupChainApp() {
	consAddress := sdk.ConsAddress(utiltx.GenerateAddress().Bytes())

	// init app
	chainID := constants.TestnetFullChainId
	suite.app = helpers.Setup(false, feemarkettypes.DefaultGenesisState(), chainID)
	header := tmproto.Header{
		Height:          1,
		ChainID:         chainID,
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}

	suite.ctx = suite.app.BaseApp.NewContext(false).WithBlockHeader(header).WithChainID(chainID)

	{
		// Finalize & commit block so the query context can be created.
		// The query context is needed for the signing texture to query coin metadata.
		_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:             header.Height,
			Hash:               header.AppHash,
			Time:               header.Time,
			ProposerAddress:    header.ProposerAddress,
			NextValidatorsHash: header.NextValidatorsHash,
		})
		suite.Require().NoError(err)
		_, err = suite.app.Commit()
		suite.Require().NoError(err)

		header.Height++
		header.Time = header.Time.Add(time.Second)
		suite.ctx = suite.ctx.
			WithBlockHeader(header).
			WithMultiStore(suite.app.CommitMultiStore().CacheMultiStore())
	}
}

func (suite *LedgerTestSuite) NewKeyringAndCtxs(krHome string, input io.Reader, encCfg params.EncodingConfig) (keyring.Keyring, client.Context, context.Context) {
	kr, err := keyring.New(
		sdk.KeyringServiceName(),
		keyring.BackendTest,
		krHome,
		input,
		encCfg.Codec,
		suite.MockKeyringOption(),
	)
	suite.Require().NoError(err)
	suite.accRetriever = mocks.NewAccountRetriever(suite.T())

	initClientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		// NOTE: cmd.Execute() panics without account retriever
		WithAccountRetriever(suite.accRetriever).
		WithTxConfig(encCfg.TxConfig).
		WithLedgerHasProtobuf(true).
		WithUseLedger(true).
		WithKeyring(kr).
		WithClient(mocks.MockCometBftRPC{Client: rpcclientmock.Client{}}).
		WithChainID(constants.TestnetFullChainId)

	ctx := context.Background()

	srvCtx := server.NewDefaultContext()
	ctx = context.WithValue(ctx, server.ServerContextKey, srvCtx)

	{ // create a new tx config with textual signing enabled
		txConfigWithTextual, err := utils.GetTxConfigWithSignModeTextureEnabled(
			authtxconfig.NewBankKeeperCoinMetadataQueryFn(s.app.BankKeeper),
			initClientCtx.Codec,
		)
		suite.Require().NoError(err)
		initClientCtx = initClientCtx.WithTxConfig(txConfigWithTextual)
		ctx = context.WithValue(ctx, client.ClientContextKey, &initClientCtx)
	}

	{ // inject the query context into the context so signing texture can query coin metadata
		queryCtx, err := suite.app.CreateQueryContext(suite.app.LastBlockHeight(), false)
		suite.Require().NoError(err)
		ctx = context.WithValue(ctx, sdk.SdkContextKey, queryCtx)
	}

	{ // update the client context with the cmd context so signing texture can use the context to query coin metadata
		initClientCtx = initClientCtx.WithCmdContext(ctx)
		ctx = context.WithValue(ctx, client.ClientContextKey, &initClientCtx)
	}

	return kr, initClientCtx, ctx
}

func (suite *LedgerTestSuite) addKeyCmd() *cobra.Command {
	cmd := keys.AddKeyCommand()

	algoFlag := cmd.Flag(flags.FlagKeyType)
	algoFlag.DefValue = string(hd.EthSecp256k1Type)

	err := algoFlag.Value.Set(string(hd.EthSecp256k1Type))
	suite.Require().NoError(err)

	cmd.Flags().AddFlagSet(keys.Commands().PersistentFlags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd).WithKeyringOptions(hd.MultiSecp256k1Option())
		clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}
		buf := bufio.NewReader(clientCtx.Input)
		return clientkeys.RunAddCmd(clientCtx, cmd, args, buf)
	}
	return cmd
}

func (suite *LedgerTestSuite) MockKeyringOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = appkeyring.SupportedAlgorithms
		options.SupportedAlgosLedger = appkeyring.SupportedAlgorithmsLedger
		options.LedgerDerivation = func() (cosmosledger.SECP256K1, error) { return suite.ledger, nil }
		options.LedgerCreateKey = appkeyring.CreatePubkey
		options.LedgerAppName = appkeyring.LedgerAppName
		options.LedgerSigSkipDERConv = appkeyring.SkipDERConversion
	}
}

func (suite *LedgerTestSuite) FormatFlag(flag string) string {
	return fmt.Sprintf("--%s", flag)
}
