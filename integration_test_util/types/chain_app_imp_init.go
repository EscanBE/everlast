package types

//goland:noinspection SpellCheckingInspection
import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cpctypes "github.com/EscanBE/everlast/x/cpc/types"

	sdkdb "github.com/cosmos/cosmos-db"
	sdkserver "github.com/cosmos/cosmos-sdk/server"

	"github.com/EscanBE/everlast/app/params"

	cmtdb "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	chainapp "github.com/EscanBE/everlast/app"
	"github.com/EscanBE/everlast/constants"
	itutilutils "github.com/EscanBE/everlast/integration_test_util/utils"
	evmtypes "github.com/EscanBE/everlast/x/evm/types"
	feemarkettypes "github.com/EscanBE/everlast/x/feemarket/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var defaultConsensusParams = &cmttypes.ConsensusParams{
	Block: cmttypes.BlockParams{
		MaxBytes: 200000,
		MaxGas:   40000000, // 40m
	},
	Evidence: cmttypes.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: cmttypes.ValidatorParams{
		PubKeyTypes: []string{
			cmttypes.ABCIPubKeyTypeEd25519,
		},
	},
}

const CometBFTGovVotingPeriod = 5 * time.Second

func NewChainApp(chainCfg ChainConfig, disableCometBFT bool, testConfig TestConfig, encCfg params.EncodingConfig, db cmtdb.DB, validatorAccounts TestAccounts, walletAccounts TestAccounts, genesisAccountBalance sdk.Coins, tempHolder *TemporaryHolder, logger log.Logger) (chainApp ChainApp, cometApp CometBftApp, validatorSet *cmttypes.ValidatorSet) {
	defaultNodeHome := chainapp.DefaultNodeHome
	moduleBasics := chainapp.ModuleBasics

	// create validator set
	var validators []*cmttypes.Validator
	for _, validatorAccount := range validatorAccounts {
		//goland:noinspection GoDeprecation
		pv := mock.PV{
			PrivKey: &cosmosed25519.PrivKey{ //nolint:staticcheck
				Key: ed25519.NewKeyFromSeed(validatorAccount.PrivateKey.Key),
			},
		}
		pubKey, err := pv.GetPubKey()
		if err != nil {
			panic(err)
		}
		validators = append(validators, cmttypes.NewValidator(pubKey, 1))
	}
	valSet := cmttypes.NewValidatorSet(validators)

	// generate genesis accounts
	var genesisValidatorAccounts []authtypes.GenesisAccount
	var genesisWalletAccounts []authtypes.GenesisAccount
	var genesisBalances []banktypes.Balance
	var signingInfos []slashingtypes.SigningInfo
	for i, account := range append(validatorAccounts, walletAccounts...) {
		acc := authtypes.NewBaseAccount(account.GetCosmosAddress(), account.GetPubKey(), uint64(i), 0)

		switch account.Type {
		case TestAccountTypeValidator:
			genesisValidatorAccounts = append(genesisValidatorAccounts, acc)

			signingInfos = append(signingInfos, slashingtypes.SigningInfo{
				Address: account.GetConsensusAddress().String(),
				ValidatorSigningInfo: slashingtypes.ValidatorSigningInfo{
					Address:             account.GetConsensusAddress().String(),
					StartHeight:         0,
					IndexOffset:         0,
					JailedUntil:         time.Time{},
					Tombstoned:          false,
					MissedBlocksCounter: 0,
				},
			})

			break
		case TestAccountTypeWallet:
			genesisWalletAccounts = append(genesisWalletAccounts, acc)

			break
		default:
			panic(fmt.Sprintf("unknown account type %d", account.Type))
		}

		genesisBalances = append(genesisBalances, banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   genesisAccountBalance,
		})
	}

	app := chainapp.NewEverLast(
		logger,           // logger
		sdkdb.NewMemDB(), // db
		nil,              // trace store
		true,             // load latest
		map[int64]bool{}, // skipUpgradeHeights
		defaultNodeHome,  // homePath
		0,                // invCheckPeriod
		encCfg,           // encodingConfig
		simtestutil.NewAppOptionsWithFlagHome(defaultNodeHome), // appOpts
		baseapp.SetChainID(chainCfg.CosmosChainId),             // baseAppOptions
	)

	// init chain must be called to stop deliverState from being nil
	genesisState := moduleBasics.DefaultGenesis(encCfg.Codec)

	genesisState = genesisStateWithValSet(chainCfg, disableCometBFT, testConfig, encCfg.Codec, genesisState, valSet, genesisValidatorAccounts, genesisWalletAccounts, genesisBalances, signingInfos)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	cai := &chainAppImp{
		app: app,
	}

	genesisDoc := cmttypes.GenesisDoc{
		GenesisTime:     time.Time{},
		ChainID:         chainCfg.CosmosChainId,
		InitialHeight:   0,
		ConsensusParams: defaultConsensusParams,
		Validators:      make([]cmttypes.GenesisValidator, len(valSet.Validators)),
		AppHash:         nil,
		AppState:        stateBytes,
	}

	for i, validator := range valSet.Validators {
		genesisDoc.Validators[i] = cmttypes.GenesisValidator{
			Address: validator.Address,
			PubKey:  validator.PubKey,
			Power:   validator.VotingPower,
			Name:    "",
		}
	}
	tempHolder.CacheGenesisDoc(&genesisDoc)

	if disableCometBFT {
		consensusParams := defaultConsensusParams.ToProto()
		_, err := app.InitChain(&abci.RequestInitChain{
			ChainId:         chainCfg.CosmosChainId,
			ConsensusParams: &consensusParams,
			Validators:      []abci.ValidatorUpdate{},
			AppStateBytes:   stateBytes,
			InitialHeight:   0,
		})
		if err != nil {
			panic(err)
		}
		cometApp = nil
	} else {
		validator := validatorAccounts.Number(1)
		if validator.GetValidatorAddress().String() != sdk.ValAddress(validator.GetPubKey().Address()).String() {
			panic("validator address does not match")
		}
		node, rpcPort, tempFiles := itutilutils.StartCometBFTNode(sdkserver.NewCometABCIWrapper(app), &genesisDoc, db, validator.GetTmPrivKey(), logger)
		for _, tempFile := range tempFiles {
			tempHolder.AddTempFile(tempFile)
		}
		cometApp = NewCometBftApp(node, rpcPort)
	}

	return cai, cometApp, valSet
}

func genesisStateWithValSet(chainCfg ChainConfig, disableCometBFT bool, testConfig TestConfig, codec codec.Codec, genesisState chainapp.GenesisState, valSet *cmttypes.ValidatorSet, genesisValidatorAccounts []authtypes.GenesisAccount, genesisWalletAccounts []authtypes.GenesisAccount, balances []banktypes.Balance, signingInfos []slashingtypes.SigningInfo) chainapp.GenesisState {
	genesisAccounts := append(genesisValidatorAccounts, genesisWalletAccounts...)

	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genesisAccounts)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for i, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			panic(err)
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			panic(err)
		}
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.OneInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genesisValidatorAccounts[i].GetAddress().String(), sdk.ValAddress(val.Address).String(), sdkmath.LegacyOneDec()))

		totalSupply = totalSupply.Add(sdk.NewCoin(chainCfg.BaseDenom, bondAmt))
	}

	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = chainCfg.BaseDenom
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(chainCfg.BaseDenom, bondAmt.MulRaw(int64(len(validators))))},
	})

	// update total supply
	baseDenomDisplay := strings.ToUpper(chainCfg.BaseDenom[1:])
	denomMetadata := []banktypes.Metadata{
		{
			Description: "Base denom metadata",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    chainCfg.BaseDenom,
					Exponent: 0,
				},
				{
					Denom:    baseDenomDisplay,
					Exponent: constants.BaseDenomExponent,
				},
			},
			Base:    chainCfg.BaseDenom,
			Display: baseDenomDisplay,
			Name:    baseDenomDisplay,
			Symbol:  baseDenomDisplay,
		},
	}
	for _, secondaryDenomUnit := range testConfig.SecondaryDenomUnits {
		secondDenomDisplay := strings.ToUpper(secondaryDenomUnit.Denom[1:])
		denomMetadata = append(denomMetadata, banktypes.Metadata{
			Description: "Second denom metadata",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    secondaryDenomUnit.Denom,
					Exponent: 0,
				},
				{
					Denom:    secondDenomDisplay,
					Exponent: secondaryDenomUnit.Exponent,
				},
			},
			Base:    secondaryDenomUnit.Denom,
			Display: secondDenomDisplay,
			Name:    secondDenomDisplay,
			Symbol:  secondDenomDisplay,
		},
		)
	}

	{
		bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, denomMetadata, []banktypes.SendEnabled{})
		genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)
	}

	{
		// x/feemarket
		feeMarketGenesis := feemarkettypes.DefaultGenesisState()
		if feeMarketGenesis != nil {
			genesisState[feemarkettypes.ModuleName] = codec.MustMarshalJSON(feeMarketGenesis)
		}
	}

	{
		// x/evm
		var evmGenesis *evmtypes.GenesisState
		evmGenesis = evmtypes.DefaultGenesisState()
		if evmGenesis != nil {
			evmGenesis.Params.EvmDenom = chainCfg.BaseDenom
			genesisState[evmtypes.ModuleName] = codec.MustMarshalJSON(evmGenesis)
		}
	}

	{
		// x/gov
		var govGenesis *govv1types.GenesisState
		govGenesis = govv1types.DefaultGenesisState()
		if govGenesis != nil {
			govGenesis.Params.MinDeposit[0].Denom = chainCfg.BaseDenom
			govGenesis.Params.MinDeposit[0].Amount = sdkmath.NewIntFromUint64(2)
			var votingPeriod time.Duration
			if disableCometBFT {
				votingPeriod = 30 * time.Minute
			} else {
				// due to CometBFT block time not configurable time jumping, we need to set a low voting period
				votingPeriod = CometBFTGovVotingPeriod
			}
			govGenesis.Params.VotingPeriod = &votingPeriod
			genesisState[govtypes.ModuleName] = codec.MustMarshalJSON(govGenesis)
		}
	}

	{
		// x/mint
		var mintGenesis *minttypes.GenesisState
		mintGenesis = minttypes.DefaultGenesisState()
		if mintGenesis != nil {
			mintGenesis.Params.MintDenom = chainCfg.BaseDenom
			genesisState[minttypes.ModuleName] = codec.MustMarshalJSON(mintGenesis)
		}
	}

	{
		// x/slashing
		var slashingGenesis *slashingtypes.GenesisState
		slashingGenesis = slashingtypes.DefaultGenesisState()
		if slashingGenesis != nil {
			slashingGenesis.SigningInfos = signingInfos
			genesisState[slashingtypes.ModuleName] = codec.MustMarshalJSON(slashingGenesis)
		}
	}

	{
		// x/cpc
		var cpcGenesis *cpctypes.GenesisState
		cpcGenesis = cpctypes.DefaultGenesis()
		if cpcGenesis != nil {
			cpcGenesis.DeployErc20Native = false
			genesisState[cpctypes.ModuleName] = codec.MustMarshalJSON(cpcGenesis)
		}
	}

	return genesisState
}
