package ibc

import (
	"testing"

	cmdcfg "github.com/EscanBE/everlast/v12/cmd/config"

	sdkmath "cosmossdk.io/math"

	"github.com/EscanBE/everlast/v12/constants"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	teststypes "github.com/EscanBE/everlast/v12/types/tests"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

func init() {
	cfg := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(cfg)
	cmdcfg.SetBip44CoinType(cfg)
}

func TestGetTransferSenderRecipient(t *testing.T) {
	testCases := []struct {
		name         string
		packet       channeltypes.Packet
		expSender    string
		expRecipient string
		expError     bool
	}{
		{
			name:         "fail - empty packet",
			packet:       channeltypes.Packet{},
			expSender:    "",
			expRecipient: "",
			expError:     true,
		},
		{
			name: "fail - invalid packet data",
			packet: channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			expSender:    "",
			expRecipient: "",
			expError:     true,
		},
		{
			name: "fail - empty FungibleTokenPacketData",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{},
				),
			},
			expSender:    "",
			expRecipient: "",
			expError:     true,
		},
		{
			name: "fail - invalid sender",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "123456",
					},
				),
			},
			expSender:    "",
			expRecipient: "",
			expError:     true,
		},
		{
			name: "fail - invalid recipient",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: constants.Bech32Prefix + "1",
						Amount:   "123456",
					},
				),
			},
			expSender:    "",
			expRecipient: "",
			expError:     true,
		},
		{
			name: "pass - valid cosmos sender, everlast recipient",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "123456",
					},
				),
			},
			expSender:    "evl1qql8ag4cluz6r4dz28p3w00dnc9w8ueuuv648f",
			expRecipient: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
			expError:     false,
		},
		{
			name: "pass - valid everlast sender, cosmos recipient",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Amount:   "123456",
					},
				),
			},
			expSender:    "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
			expRecipient: "evl1qql8ag4cluz6r4dz28p3w00dnc9w8ueuuv648f",
			expError:     false,
		},
		{
			name: "pass - valid osmosis sender, everlast recipient",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "123456",
					},
				),
			},
			expSender:    "evl1qql8ag4cluz6r4dz28p3w00dnc9w8ueuuv648f",
			expRecipient: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
			expError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sender, recipient, _, _, err := GetTransferSenderRecipient(tc.packet)
			if tc.expError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expSender, sender.String())
			require.Equal(t, tc.expRecipient, recipient.String())
		})
	}
}

func TestGetTransferAmount(t *testing.T) {
	testCases := []struct {
		name      string
		packet    channeltypes.Packet
		expAmount string
		expError  bool
	}{
		{
			name:      "fail - empty packet",
			packet:    channeltypes.Packet{},
			expAmount: "",
			expError:  true,
		},
		{
			name: "fail - invalid packet data",
			packet: channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			expAmount: "",
			expError:  true,
		},
		{
			name: "fail - invalid amount - empty",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "",
					},
				),
			},
			expAmount: "",
			expError:  true,
		},
		{
			name: "fail - invalid amount - non-int",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "test",
					},
				),
			},
			expAmount: "test",
			expError:  true,
		},
		{
			name: "pass - valid",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "10000",
					},
				),
			},
			expAmount: "10000",
			expError:  false,
		},
		{
			name: "pass - valid - any amount",
			packet: channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "evl1x2w87cvt5mqjncav4lxy8yfreynn273xj3szj4",
						Amount:   "1",
					},
				),
			},
			expAmount: "1",
			expError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			amt, err := GetTransferAmount(tc.packet)
			if tc.expError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expAmount, amt)
		})
	}
}

func TestGetReceivedCoin(t *testing.T) {
	testCases := []struct {
		name       string
		srcPort    string
		srcChannel string
		dstPort    string
		dstChannel string
		rawDenom   string
		rawAmount  string
		expCoin    sdk.Coin
	}{
		{
			name:       "transfer unwrapped coin to destination which is not its source",
			srcPort:    "transfer",
			srcChannel: "channel-0",
			dstPort:    "transfer",
			dstChannel: "channel-0",
			rawDenom:   "uosmo",
			rawAmount:  "10",
			expCoin:    sdk.Coin{Denom: teststypes.UosmoIbcdenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:       "transfer ibc wrapped coin to destination which is its source",
			srcPort:    "transfer",
			srcChannel: "channel-0",
			dstPort:    "transfer",
			dstChannel: "channel-0",
			rawDenom:   "transfer/channel-0/" + constants.BaseDenom,
			rawAmount:  "10",
			expCoin:    sdk.Coin{Denom: constants.BaseDenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:       "transfer 2x ibc wrapped coin to destination which is its source",
			srcPort:    "transfer",
			srcChannel: "channel-0",
			dstPort:    "transfer",
			dstChannel: "channel-2",
			rawDenom:   "transfer/channel-0/transfer/channel-1/uatom",
			rawAmount:  "10",
			expCoin:    sdk.Coin{Denom: teststypes.UatomIbcdenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:       "transfer ibc wrapped coin to destination which is not its source",
			srcPort:    "transfer",
			srcChannel: "channel-0",
			dstPort:    "transfer",
			dstChannel: "channel-0",
			rawDenom:   "transfer/channel-1/uatom",
			rawAmount:  "10",
			expCoin:    sdk.Coin{Denom: teststypes.UatomOsmoIbcdenom, Amount: sdkmath.NewInt(10)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coin := GetReceivedCoin(tc.srcPort, tc.srcChannel, tc.dstPort, tc.dstChannel, tc.rawDenom, tc.rawAmount)
			require.Equal(t, tc.expCoin, coin)
		})
	}
}

func TestGetSentCoin(t *testing.T) {
	testCases := []struct {
		name      string
		rawDenom  string
		rawAmount string
		expCoin   sdk.Coin
	}{
		{
			name:      "get unwrapped native coin",
			rawDenom:  constants.BaseDenom,
			rawAmount: "10",
			expCoin:   sdk.Coin{Denom: constants.BaseDenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:      "get ibc wrapped native coin",
			rawDenom:  "transfer/channel-0/" + constants.BaseDenom,
			rawAmount: "10",
			expCoin:   sdk.Coin{Denom: teststypes.NativeCoinIbcdenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:      "get ibc wrapped uosmo coin",
			rawDenom:  "transfer/channel-0/uosmo",
			rawAmount: "10",
			expCoin:   sdk.Coin{Denom: teststypes.UosmoIbcdenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:      "get ibc wrapped uatom coin",
			rawDenom:  "transfer/channel-1/uatom",
			rawAmount: "10",
			expCoin:   sdk.Coin{Denom: teststypes.UatomIbcdenom, Amount: sdkmath.NewInt(10)},
		},
		{
			name:      "get 2x ibc wrapped uatom coin",
			rawDenom:  "transfer/channel-0/transfer/channel-1/uatom",
			rawAmount: "10",
			expCoin:   sdk.Coin{Denom: teststypes.UatomOsmoIbcdenom, Amount: sdkmath.NewInt(10)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coin := GetSentCoin(tc.rawDenom, tc.rawAmount)
			require.Equal(t, tc.expCoin, coin)
		})
	}
}
