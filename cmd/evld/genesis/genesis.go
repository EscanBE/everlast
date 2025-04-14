package genesis

import (
	"github.com/spf13/cobra"
)

// Cmd creates a main CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis",
		Short: "Allow custom genesis",
	}

	cmd.AddCommand(
		NewImproveGenesisCmd(),
	)

	return cmd
}
