package commands

import (
	"fmt"
	"os"

	"github.com/johnmccabe/go-vmpooler/vm"
	"github.com/johnmccabe/vmpooler-bitbar/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:    "delete",
	Run:    runDelete,
	Hidden: true,
}

func runDelete(cmd *cobra.Command, args []string) {
	cfg, err := config.Read()
	if err != nil {
		os.Exit(1)
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	target := args[0]

	vmclient := vm.NewClient(cfg.Endpoint, cfg.Token)

	if target != "all" {
		err = vmclient.Delete(target)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		vms, err := vmclient.GetAll()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		for _, vm := range vms {
			err = vmclient.Delete(vm.Hostname)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}
