package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/spf13/cast"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtos "github.com/cometbft/cometbft/libs/os"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdkdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	"github.com/EscanBE/everlast/app/antedl"
	"github.com/EscanBE/everlast/app/antedl/duallane"
	"github.com/EscanBE/everlast/app/keepers"
	"github.com/EscanBE/everlast/app/params"
	"github.com/EscanBE/everlast/app/upgrades"
	"github.com/EscanBE/everlast/client/docs"
	"github.com/EscanBE/everlast/constants"
	"github.com/EscanBE/everlast/ethereum/eip712"
	evertypes "github.com/EscanBE/everlast/types"
	"github.com/EscanBE/everlast/utils"

	// Force-load the tracer engines to trigger registration due to Go-Ethereum v1.10.15 changes
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	Upgrades  []upgrades.Upgrade
	HardForks []upgrades.Fork
)

var (
	_ servertypes.Application = (*EverLast)(nil)
	_ ibctesting.TestingApp   = (*EverLast)(nil)
)

// EverLast implements an extended ABCI application.
// It is an application that may process transactions
// through Ethereum's EVM running atop of CometBFT consensus.
type EverLast struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	// encoding
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	invCheckPeriod uint

	// the module manager
	mm           *module.Manager
	ModuleBasics module.BasicManager // delivered from module manager, with installed codec

	// the configurator
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, constants.ApplicationHome)

	sdk.DefaultPowerReduction = evertypes.PowerReduction // 10^18
}

// NewEverLast returns a reference to a new initialized EverLast application.
func NewEverLast(
	logger log.Logger,
	db sdkdb.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig params.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *EverLast {
	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	defer func() {
		eip712.SetEncodingConfig(encodingConfig)
	}()

	// App Opts
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE we use custom transaction decoder that supports the sdk.Tx interface instead of sdk.StdTx
	baseApp := baseapp.NewBaseApp(
		constants.ApplicationName,
		logger,
		db,
		txConfig.TxDecoder(),
		baseAppOptions...,
	)

	baseApp.SetCommitMultiStoreTracer(traceStore)
	baseApp.SetVersion(version.Version)
	baseApp.SetInterfaceRegistry(interfaceRegistry)
	baseApp.SetTxEncoder(txConfig.TxEncoder())

	chainApp := &EverLast{
		BaseApp:           baseApp,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}

	moduleAccountAddresses := chainApp.ModuleAccountAddrs()

	// Setup keepers
	chainApp.AppKeepers = keepers.NewAppKeeper(
		appCodec,
		baseApp,
		legacyAmino,
		maccPerms,
		moduleAccountAddresses,
		chainApp.BlockedModuleAccountAddrs(moduleAccountAddresses),
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		logger,
		appOpts,
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	chainApp.mm = module.NewManager(appModules(chainApp, encodingConfig, skipGenesisInvariants)...)
	chainApp.ModuleBasics = newBasicManagerFromManager(chainApp)

	{
		txConfigWithTextual, err := utils.GetTxConfigWithSignModeTextureEnabled(
			authtxconfig.NewBankKeeperCoinMetadataQueryFn(chainApp.BankKeeper),
			appCodec,
		)
		if err != nil {
			panic(err)
		}
		chainApp.txConfig = txConfigWithTextual
		encodingConfig.TxConfig = txConfigWithTextual
	}

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	chainApp.mm.SetOrderBeginBlockers(orderBeginBlockers()...)

	chainApp.mm.SetOrderEndBlockers(orderEndBlockers()...)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	chainApp.mm.SetOrderInitGenesis(orderInitBlockers()...)

	chainApp.mm.RegisterInvariants(chainApp.CrisisKeeper)
	chainApp.configurator = module.NewConfigurator(chainApp.appCodec, chainApp.MsgServiceRouter(), chainApp.GRPCQueryRouter())
	if err := chainApp.mm.RegisterServices(chainApp.configurator); err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(chainApp.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(chainApp.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(chainApp.GRPCQueryRouter(), reflectionSvc)
	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(chainApp.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// initialize stores
	chainApp.MountKVStores(chainApp.GetKVStoreKey())
	chainApp.MountTransientStores(chainApp.GetTransientStoreKey())
	chainApp.MountMemoryStores(chainApp.GetMemoryStoreKey())

	// chainApp.setAnteHandler(txConfig)
	chainApp.setDualLaneAnteHandler(txConfig)
	chainApp.setPostHandler()

	chainApp.SetInitChainer(chainApp.InitChainer)
	chainApp.SetBeginBlocker(chainApp.BeginBlocker)
	chainApp.SetEndBlocker(chainApp.EndBlocker)

	chainApp.setupUpgradeHandlers()
	chainApp.setupUpgradeStoreLoaders()

	if loadLatest {
		if err := chainApp.LoadLatestVersion(); err != nil {
			cmtos.Exit(err.Error())
		}
	}

	return chainApp
}

// Name returns the name of the App
func (app *EverLast) Name() string { return app.BaseApp.Name() }

// BeginBlocker runs the CometBFT ABCI BeginBlock logic. It executes state changes at the beginning
// of the new block for every registered module. If there is a registered fork at the current height,
// BeginBlocker will schedule the upgrade plan and perform the state migration (if any).
func (app *EverLast) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	// Perform any scheduled forks before executing the modules logic
	app.scheduleForkUpgrade(ctx)
	return app.mm.BeginBlock(ctx)
}

// EndBlocker updates every end block
func (app *EverLast) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// InitChainer updates at chain initialization
func (app *EverLast) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap()); err != nil {
		panic(err)
	}

	response, err := app.mm.InitGenesis(ctx, app.appCodec, genesisState)
	if err != nil {
		panic(err)
	}

	return response, nil
}

// LoadHeight loads state at a particular height
func (app *EverLast) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *EverLast) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *EverLast) BlockedModuleAccountAddrs(modAccAddrs map[string]bool) map[string]bool {
	blockedAddrs := make(map[string]bool)

	for acc := range modAccAddrs {
		blockedAddrs[acc] = true
	}

	return blockedAddrs
}

// LegacyAmino returns EverLast's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *EverLast) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns EverLast's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *EverLast) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns EverLast's InterfaceRegistry
func (app *EverLast) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *EverLast) RegisterAPIRoutes(apiSvr *api.Server, apiConfig srvconfig.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new CometBFT queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	node.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := RegisterSwaggerAPI(clientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterNodeService allows query minimum-gas-prices in app.toml
func (app *EverLast) RegisterNodeService(clientCtx client.Context, cfg srvconfig.Config) {
	node.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *EverLast) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *EverLast) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *EverLast) setDualLaneAnteHandler(txConfig client.TxConfig) {
	options := antedl.HandlerOptions{
		Cdc:                    app.appCodec,
		AccountKeeper:          &app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		ExtensionOptionChecker: evertypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              app.EvmKeeper,
		VAuthKeeper:            &app.VAuthKeeper,
		StakingKeeper:          app.StakingKeeper,
		FeegrantKeeper:         &app.FeeGrantKeeper,
		DistributionKeeper:     &app.DistrKeeper,
		IBCKeeper:              app.IBCKeeper,
		FeeMarketKeeper:        &app.FeeMarketKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		SigGasConsumer:         duallane.SigVerificationGasConsumer,
		TxFeeChecker:           duallane.DualLaneFeeChecker(app.EvmKeeper, app.FeeMarketKeeper),
	}.WithDefaultDisabledNestedMsgs()

	if err := options.Validate(); err != nil {
		panic(err)
	}

	app.SetAnteHandler(antedl.NewAnteHandler(options))
}

func (app *EverLast) setPostHandler() {
	postHandler, err := NewPostHandler()
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

func (app *EverLast) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName, // Sample v13.0.0 upgrade handler
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				&app.AppKeepers,
			),
		)
	}
}

func (app *EverLast) setupUpgradeStoreLoaders() {
	// When a planned update height is reached, the old binary will panic
	// writing on disk the height and name of the update that triggered it
	// This will read that value, and execute the preparations for the upgrade.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrade := range Upgrades {
		upgrade := upgrade
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}
}

// IBC Go TestingApp functions

// GetBaseApp implements the TestingApp interface.
func (app *EverLast) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetTxConfig implements the TestingApp interface.
func (app *EverLast) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *EverLast) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// GetStakingKeeper implements the TestingApp interface.
func (app *EverLast) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetStakingKeeperSDK implements the TestingApp interface.
func (app *EverLast) GetStakingKeeperSDK() stakingkeeper.Keeper {
	return *app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *EverLast) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *EverLast) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router, enableSwagger bool) error {
	if !enableSwagger {
		return nil
	}

	root, err := fs.Sub(docs.SwaggerUI, "swagger-ui")
	if err != nil {
		return err
	}

	staticServer := http.FileServer(http.FS(root))
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))

	return nil
}
