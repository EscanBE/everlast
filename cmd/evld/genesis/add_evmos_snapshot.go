package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/cosmos/cosmos-sdk/types/bech32"

	sdkmath "cosmossdk.io/math"

	"github.com/EscanBE/everlast/constants"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func NewAddEvmosSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-evmos-snapshot",
		Short: "Add Evmos balance from snapshot to the genesis file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generalGenesisUpdateFunc(cmd, func(genesis map[string]json.RawMessage, clientCtx client.Context) error {
				{ // Update the app state
					var appState map[string]json.RawMessage
					err := json.Unmarshal(genesis["app_state"], &appState)
					if err != nil {
						return fmt.Errorf("failed to unmarshal app state: %w", err)
					}

					codec := clientCtx.Codec

					// Update bank genesis state
					var bankGenesisState banktypes.GenesisState
					codec.MustUnmarshalJSON(appState["bank"], &bankGenesisState)

					addBankDataFromEvmosSnapshot(&bankGenesisState)

					appState["bank"] = codec.MustMarshalJSON(&bankGenesisState)

					// Marshal the updated app state back to genesis
					updatedAppState, err := json.Marshal(appState)
					if err != nil {
						return fmt.Errorf("failed to marshal updated app state: %w", err)
					}
					genesis["app_state"] = updatedAppState
				}

				return nil
			})
		},
	}

	return cmd
}

func addBankDataFromEvmosSnapshot(
	bankGenesisState *banktypes.GenesisState,
) {
	bankBalanceByAddress := make(map[string]sdk.Coins)
	expectedOriginalTotalSupply := sdkmath.NewInt(0)
	for _, balance := range bankGenesisState.Balances {
		bankBalanceByAddress[balance.Address] = balance.Coins
		expectedOriginalTotalSupply = expectedOriginalTotalSupply.Add(balance.Coins.AmountOf(constants.BaseDenom))
	}

	// Assert the original total supply is equal to the sum of all original balances as we expect that.
	if !expectedOriginalTotalSupply.Equal(bankGenesisState.Supply.AmountOf(constants.BaseDenom)) {
		panic(fmt.Sprintf("original total supply %s does not match the sum of balances %s", bankGenesisState.Supply.AmountOf(constants.BaseDenom).String(), expectedOriginalTotalSupply.String()))
	}

	defer func() {
		totalSupplyOfNativeLater := sdkmath.ZeroInt()

		// complete the bank genesis state after updated balances for accounts

		// copy the modified balance back to the bank genesis state
		bankGenesisState.Balances = make([]banktypes.Balance, 0, len(bankBalanceByAddress))
		for addr, balance := range bankBalanceByAddress {
			if balance.IsZero() {
				continue
			}
			bankGenesisState.Balances = append(bankGenesisState.Balances, banktypes.Balance{
				Address: addr,
				Coins:   balance,
			})
			totalSupplyOfNativeLater = totalSupplyOfNativeLater.Add(balance.AmountOf(constants.BaseDenom))
		}

		// update the total supply
		newTotalSupply := sdk.Coins{sdk.NewCoin(constants.BaseDenom, totalSupplyOfNativeLater)}
		for _, coin := range bankGenesisState.Supply {
			if coin.Denom == constants.BaseDenom {
				// skip because we added just above
				continue
			}
			newTotalSupply = newTotalSupply.Add(coin)
		}
		bankGenesisState.Supply = newTotalSupply

		toDisplayAmount := func(amount sdkmath.Int) sdkmath.Int {
			return amount.Quo(sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(constants.BaseDenomExponent), nil)))
		}

		fmt.Println("Total supply of native token after adding evmos snapshot:", toDisplayAmount(totalSupplyOfNativeLater))
		fmt.Println("Increased by:", toDisplayAmount(totalSupplyOfNativeLater.Sub(expectedOriginalTotalSupply)))
	}()

	{ // check directory
		if _, err := os.Stat("Makefile"); err != nil {
			panic(fmt.Errorf("not in the root directory of project, required to access the snapshot file: %w", err))
		}
	}

	const evmosSnapshotFile1 = "cmd/evld/genesis/snapshot_21302452.json"
	const evmosSnapshotFile2 = "cmd/evld/genesis/snapshot_28318578.json"
	const magicNumberForBalanceFile1 float64 = 23.272
	const magicNumberForBalanceFile2 float64 = 299.043
	const keepRatioEachSnapshot float64 = 0.5 // two snapshots so keep half each

	type userBalance struct {
		Address string `json:"a"`
		Balance string `json:"b"`
	}

	readSnapshotFile := func(filePath string) ([]userBalance, error) {
		bzSnapshot, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var content map[string][]userBalance
		if err := json.Unmarshal(bzSnapshot, &content); err != nil {
			return nil, err
		}

		balances := content["bank"]
		if len(balances) < 1 {
			return nil, fmt.Errorf("no balances found in the snapshot file")
		}

		return balances, nil
	}

	balancesFromSnapshot1, err := readSnapshotFile(evmosSnapshotFile1)
	if err != nil {
		panic(fmt.Errorf("failed to read evmos snapshot file 1: %w", err))
	}

	balancesFromSnapshot2, err := readSnapshotFile(evmosSnapshotFile2)
	if err != nil {
		panic(fmt.Errorf("failed to read evmos snapshot file 2: %w", err))
	}

	// exponent of Evmos was reduced to 6 during snapshot so we restore it to 18
	amountMultiplier := sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(constants.BaseDenomExponent-constants.SnapshotEvmosExponent), nil))

	processBalancesFromSnapshot := func(balances []userBalance, magicNumber float64) {
		for _, balance := range balancesFromSnapshot1 {
			rawAmount, ok := sdkmath.NewIntFromString(balance.Balance)
			if !ok {
				panic(fmt.Errorf("invalid balance %s of %s", balance.Balance, balance.Address))
			}

			_, bz, err := bech32.DecodeAndConvert(balance.Address)
			if err != nil {
				panic(fmt.Errorf("failed to decode bech32 address %s: %w", balance.Address, err))
			}

			address := sdk.AccAddress(bz).String()

			originalAevmosAmount := rawAmount.Mul(amountMultiplier)

			effectiveAmount := getEffectiveAmountFromEvmosSnapshotBalance(originalAevmosAmount, magicNumber, keepRatioEachSnapshot)

			bankBalanceByAddress[address] = bankBalanceByAddress[address].Add(sdk.NewCoin(constants.BaseDenom, effectiveAmount))
		}
	}

	processBalancesFromSnapshot(balancesFromSnapshot1, magicNumberForBalanceFile1)
	processBalancesFromSnapshot(balancesFromSnapshot2, magicNumberForBalanceFile2)
}

func getEffectiveAmountFromEvmosSnapshotBalance(
	amount sdkmath.Int,
	magicNumber float64,
	keepRatio float64,
) sdkmath.Int {
	v := float64(amount.QuoRaw(1e18).Int64()) / magicNumber

	var keepV float64
	for _, config := range ratioConfigs {
		if v >= config.accumulatedValue {
			keepV += config.maxKeep
			continue
		}

		keepV += config.computeKeep(v - (config.accumulatedValue - config.stepValue))
		break
	}

	keepV *= keepRatio

	return sdkmath.NewInt(int64(keepV)).MulRaw(1e18)
}

type ratioConfig struct {
	stepValue        float64
	keepRatio        float64
	accumulatedValue float64 // compute
	maxKeep          float64 // compute
}

func (rc ratioConfig) computeKeep(v float64) float64 {
	return v * rc.keepRatio
}

var ratioConfigs []*ratioConfig

func init() {
	ratioConfigs = []*ratioConfig{
		{
			stepValue: 1000,
			keepRatio: 1,
		},
		{
			stepValue: 500,
			keepRatio: .9,
		},
		{
			stepValue: 500,
			keepRatio: .8,
		},
		{
			stepValue: 500,
			keepRatio: .7,
		},
		{
			stepValue: 500,
			keepRatio: .6,
		},
		{
			stepValue: 500,
			keepRatio: .5,
		},
		{
			stepValue: 1000,
			keepRatio: .4,
		},
		{
			stepValue: 2000,
			keepRatio: .3,
		},
		{
			stepValue: 3000,
			keepRatio: .2,
		},
		{
			stepValue: 10000,
			keepRatio: .1,
		},
		{
			stepValue: 100000,
			keepRatio: .05,
		},
		{
			stepValue: 999_999_999_999,
			keepRatio: .01,
		},
	}

	var accumulatedValue float64
	for _, config := range ratioConfigs {
		accumulatedValue += config.stepValue
		config.accumulatedValue = accumulatedValue
		config.maxKeep = config.computeKeep(config.stepValue)
	}
}
