package main

import (
	"fmt"
	"os"

	"github.com/EscanBE/everlast/constants"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"

	chainapp "github.com/EscanBE/everlast/app"
	cmdcfg "github.com/EscanBE/everlast/cmd/config"
)

func main() {
	setupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, constants.ApplicationBinaryName, chainapp.DefaultNodeHome); err != nil {
		_, _ = fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func setupConfig() {
	// set the address prefixes
	cfg := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(cfg)
	cmdcfg.SetBip44CoinType(cfg)
	// TODO fix
	// if err := cmdcfg.EnableObservability(); err != nil {
	// 	panic(err)
	// }
	cfg.Seal()
}
