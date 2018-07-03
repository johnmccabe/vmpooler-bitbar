package commands

import (
	"fmt"
	"os"

	"github.com/johnmccabe/go-vmpooler/vm"
	"github.com/johnmccabe/vmpooler-bitbar/config"
	"github.com/spf13/cobra"
)

const defaultLifetimeExtension = 2

func init() {
	rootCmd.AddCommand(extendCmd)
}

var extendCmd = &cobra.Command{
	Use:    "extend",
	Run:    runExtend,
	Hidden: true,
}

func runExtend(cmd *cobra.Command, args []string) {
	cfg, err := config.Read()
	if err != nil {
		os.Exit(1)
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	target := args[0]

	vmclient := vm.NewClient(cfg.Endpoint, cfg.Token)

	var vms []vm.VM
	var virtualmachine *vm.VM

	if target == "all" {
		vms, err = vmclient.GetAll()
	} else {
		virtualmachine, err = vmclient.Get(target)
		vms = []vm.VM{*virtualmachine}
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	for _, vm := range vms {
		_, err = vmclient.SetLifetime(vm.Hostname, int(vm.Lifetime)+defaultLifetimeExtension)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

}
