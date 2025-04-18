package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                DefaultParams(),
		DeployErc20Native:     false,
		DeployStakingContract: false,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
