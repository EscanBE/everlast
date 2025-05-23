package app

import (
	"encoding/json"

	"github.com/EscanBE/everlast/app/params"
)

// GenesisState of the blockchain is represented here as a map of raw json
// messages key by an identifier string.
// The identifier is used to determine which module genesis information belongs to,
// so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(encodingConfig params.EncodingConfig) GenesisState {
	return ModuleBasics.DefaultGenesis(encodingConfig.Codec)
}
