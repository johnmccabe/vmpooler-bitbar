package commands

import (
	"fmt"
	"os"

	"github.com/johnmccabe/go-vmpooler/vm"
	"github.com/johnmccabe/vmpooler-bitbar/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newVMCmd)
}

var newVMCmd = &cobra.Command{
	Use:    "newvm",
	Run:    runNewVMCmd,
	Hidden: true,
}

func runNewVMCmd(cmd *cobra.Command, args []string) {
	cfg, err := config.Read()
	if err != nil {
		os.Exit(1)
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	template := args[0]

	vmclient := vm.NewClient(cfg.Endpoint, cfg.Token)

	_, err = vmclient.Create(template)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
