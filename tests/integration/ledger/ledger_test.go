package ledger_test

import (
	"bytes"
	"context"

	sdkmath "cosmossdk.io/math"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	params "github.com/EscanBE/everlast/app/params"

	"github.com/EscanBE/everlast/constants"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	chainapp "github.com/EscanBE/everlast/app"
	"github.com/EscanBE/everlast/crypto/hd"
	"github.com/EscanBE/everlast/tests/integration/ledger/mocks"
	"github.com/EscanBE/everlast/testutil"
	utiltx "github.com/EscanBE/everlast/testutil/tx"

	"github.com/spf13/cobra"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdktestutilcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	. "github.com/onsi/ginkgo/v2"
)

var (
	signOkMock = func(_ []uint32, msg []byte, _ byte) ([]byte, error) {
		return s.privKey.Sign(msg)
	}

	signErrMock = func(_ []uint32, msg []byte, _ byte) ([]byte, error) {
		return nil, mocks.ErrMockedSigning
	}
)

var _ = Describe("Ledger CLI and keyring functionality: ", func() {
	var (
		receiverAccAddr sdk.AccAddress
		encodingConfig  params.EncodingConfig
		kr              keyring.Keyring
		mockedIn        sdktestutil.BufferReader
		clientCtx       client.Context
		ctx             context.Context
		cmd             *cobra.Command
		krHome          string
		keyRecord       *keyring.Record
	)

	ledgerKey := "ledger_key"
	encodingConfig = chainapp.RegisterEncodingConfig()

	s.SetupTest()
	s.SetupChainApp()

	Describe("Adding a key from ledger using the CLI", func() {
		BeforeEach(func() {
			krHome = s.T().TempDir()

			cmd = s.addKeyCmd()

			mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

			kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encodingConfig)

			mocks.MClose(s.ledger)
			mocks.MGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)
		})
		Context("with default algo", func() {
			It("should use eth_secp256k1 by default and pass", func() {
				out, err := sdktestutilcli.ExecTestCLICmd(clientCtx, cmd, []string{
					ledgerKey,
					s.FormatFlag(flags.FlagUseLedger),
				})

				s.Require().NoError(err)
				s.Require().Contains(out.String(), "name: ledger_key")

				_, err = kr.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")
			})
		})
		Context("with eth_secp256k1 algo", func() {
			It("should add the ledger key ", func() {
				out, err := sdktestutilcli.ExecTestCLICmd(clientCtx, cmd, []string{
					ledgerKey,
					s.FormatFlag(flags.FlagUseLedger),
					s.FormatFlag(flags.FlagKeyType), string(hd.EthSecp256k1Type),
				})

				s.Require().NoError(err)
				s.Require().Contains(out.String(), "name: ledger_key")

				_, err = kr.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")
			})
		})
	})
	Describe("Signing a transactions", func() {
		BeforeEach(func() {
			krHome = s.T().TempDir()

			var err error

			// create add key command
			cmd = s.addKeyCmd()

			mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)
			mocks.MGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)

			kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encodingConfig)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			cmd.SetArgs([]string{
				ledgerKey,
				s.FormatFlag(flags.FlagUseLedger),
				s.FormatFlag(flags.FlagKeyType), "eth_secp256k1",
			})
			// add ledger key for following tests
			s.Require().NoError(cmd.ExecuteContext(ctx))
			keyRecord, err = kr.Key(ledgerKey)
			s.Require().NoError(err, "can't find ledger key")
		})
		Context("perform bank send", func() {
			Context("with keyring functions calling", func() {
				BeforeEach(func() {
					s.ledger = mocks.NewSECP256K1(s.T())

					mocks.MClose(s.ledger)
					mocks.MGetPublicKeySECP256K1(s.ledger, s.pubKey)
				})
				It("should return valid signature", func() {
					mocks.MSignSECP256K1(s.ledger, signOkMock, nil)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

					msg := []byte("test message")

					signed, _, err := kr.SignByAddress(ledgerAddr, msg, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
					s.Require().NoError(err, "failed to sign message")

					valid := s.pubKey.VerifySignature(msg, signed)
					s.Require().True(valid, "invalid signature returned")
				})
				It("should raise error from ledger sign function to the top", func() {
					mocks.MSignSECP256K1(s.ledger, signErrMock, mocks.ErrMockedSigning)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retrieve ledger addr from a keyring")

					msg := []byte("test message")

					_, _, err = kr.SignByAddress(ledgerAddr, msg, signingtypes.SignMode_SIGN_MODE_TEXTUAL)

					s.Require().Error(err, "false positive result, error expected")

					s.Require().Equal(mocks.ErrMockedSigning.Error(), err.Error(), "original and returned errors are not equal")
				})
			})
			Context("with cli command", func() {
				BeforeEach(func() {
					s.ledger = mocks.NewSECP256K1(s.T())

					err := testutil.FundAccount(
						s.ctx,
						s.app.BankKeeper,
						s.accAddr,
						sdk.NewCoins(
							sdk.NewCoin(constants.BaseDenom, sdkmath.NewInt(100000000000000)),
						),
					)
					s.Require().NoError(err)

					receiverAccAddr = sdk.AccAddress(utiltx.GenerateAddress().Bytes())

					cmd = bankcli.NewSendTxCmd(s.app.AccountKeeper.AddressCodec())
					mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

					kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encodingConfig)

					// register mocked funcs
					mocks.MClose(s.ledger)
					mocks.MGetPublicKeySECP256K1(s.ledger, s.pubKey)
					mocks.MEnsureExist(s.accRetriever, nil)
					mocks.MGetAccountNumberSequence(s.accRetriever, 0, 0, nil)
				})
				It("should execute bank tx cmd", func() {
					mocks.MSignSECP256K1(s.ledger, signOkMock, nil)

					cmd.SetContext(ctx)
					cmd.SetArgs([]string{
						ledgerKey,
						receiverAccAddr.String(),
						sdk.NewCoin(constants.BaseDenom, sdkmath.NewInt(1000)).String(),
						s.FormatFlag(flags.FlagUseLedger),
						s.FormatFlag(flags.FlagSkipConfirmation),
						s.FormatFlag(flags.FlagSignMode), flags.SignModeTextual,
					})
					out := bytes.NewBufferString("")
					cmd.SetErr(out)

					err := cmd.Execute()

					s.Require().NoError(err, "can't execute cli tx command")
				})
				It("should return error from ledger device", func() {
					mocks.MSignSECP256K1(s.ledger, signErrMock, mocks.ErrMockedSigning)

					cmd.SetContext(ctx)
					cmd.SetArgs([]string{
						ledgerKey,
						receiverAccAddr.String(),
						sdk.NewCoin(constants.BaseDenom, sdkmath.NewInt(1000)).String(),
						s.FormatFlag(flags.FlagUseLedger),
						s.FormatFlag(flags.FlagSkipConfirmation),
						s.FormatFlag(flags.FlagSignMode), flags.SignModeTextual,
					})
					out := bytes.NewBufferString("")
					cmd.SetErr(out)

					err := cmd.Execute()

					s.Require().Error(err, "false positive, error expected")
					s.Require().Equal(mocks.ErrMockedSigning.Error(), err.Error())
				})
			})
		})
	})
})
