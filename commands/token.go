package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/johnmccabe/go-vmpooler/token"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	rootCmd.AddCommand(tokenCmd)
}

var tokenCmd = &cobra.Command{
	Use: "token",
	Run: runTokenCmd,
}

func runTokenCmd(cmd *cobra.Command, args []string) {
	endpoint, username, password, err := getCredentials()
	if err != nil {
		log.Fatal(err.Error())
	}

	t := token.NewClient(endpoint, username, password)

	token, err := t.Generate()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("\nToken generated: %s\n", token.Token)
}

func getCredentials() (string, string, string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter vmpooler API endpoint: ")
	scanner.Scan()
	endpoint := scanner.Text()

	fmt.Print("Enter Username: ")
	scanner.Scan()
	username := scanner.Text()

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	fmt.Println()

	return endpoint, username, password, err
}
