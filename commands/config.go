package commands

// defaults read com.matryer.BitBar pluginsDirectory

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/yaml.v2"

	"github.com/johnmccabe/vmpooler-bitbar/config"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use: "config",
	Run: runConfig,
}

func runConfig(cmd *cobra.Command, args []string) {
	cfg, _ := config.Read()
	var promptInput *config.Config
	if (config.Config{}) == cfg {
		promptInput = nil
	} else {
		promptInput = &cfg
	}

	cfg, err := promptForConfig(promptInput)
	if err != nil {
		fmt.Printf("Error gathering config: [%v]", err)
		os.Exit(1)
	}

	config.EnsureConfigDir()

	f, err := os.OpenFile(config.File(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Unable to expand homedir: [%v]", err)
		os.Exit(1)
	}

	defer f.Close()

	y, _ := yaml.Marshal(cfg)

	if _, err = f.Write(y); err != nil {
		fmt.Printf("Unable to save config: [%v]", err)
		os.Exit(1)
	}

	fmt.Println("Config successfully updated.")

	refreshPlugin()

	os.Exit(0)
}

func refreshPlugin() error {
	answers := struct {
		Refresh bool
	}{}

	var refreshQuestions = []*survey.Question{
		{
			Name: "refresh",
			Prompt: &survey.Confirm{
				Message: "Refresh plugin (window will lose focus)?",
			},
		},
	}

	err := survey.Ask(refreshQuestions, &answers)
	if err != nil {
		return err
	}

	if answers.Refresh {
		cmd := exec.Command("open", "bitbar://refreshPlugin?name=vmpooler-bitbar.*?")
		fmt.Println("Refreshing plugin...")
		err := cmd.Run()
		if err != nil {
			log.Printf("Command exited unexpectedly with error: %v", err)
		}

	}

	return nil
}

func promptForConfig(cfg *config.Config) (config.Config, error) {
	answers := struct {
		Endpoint        string
		Token           string
		LifetimeWarning string
	}{}

	err := survey.Ask(questions(cfg), &answers)
	if err != nil {
		return config.Config{}, err
	}

	result := config.Config{
		Endpoint: answers.Endpoint,
		Token:    answers.Token,
	}

	ltwi, err := strconv.Atoi(answers.LifetimeWarning)
	if err != nil {
		return config.Config{}, err
	}
	result.LifetimeWarning = ltwi

	return result, nil
}

func questions(cfg *config.Config) []*survey.Question {

	endpointInput := &survey.Input{
		Message: "Vmpooler Endpoint?",
		Help:    "Your organisations vmpooler API endpoint. For example: https://vmpooler.mycompany.net/api/v1",
	}
	if cfg != nil {
		endpointInput.Default = cfg.Endpoint
	}

	tokenInput := &survey.Input{
		Message: "Vmpooler Token?",
		Help:    "Your personal Vmpooler token. For example: kpy2fn8sgjkcbyn896yilzqxwjlfake",
	}
	if cfg != nil {
		tokenInput.Default = cfg.Token
	}

	lifetimewarningInput := &survey.Input{
		Message: "VM lifetime warning threshold?",
		Help:    "VMs with a remaining lifetime less than this value in hours will be flagged in red. For example: 1",
	}
	if cfg != nil {
		lifetimewarningInput.Default = strconv.Itoa(cfg.LifetimeWarning)
	}

	questions := []*survey.Question{
		{
			Name:     "endpoint",
			Prompt:   endpointInput,
			Validate: survey.Required,
		},
		{
			Name:     "token",
			Prompt:   tokenInput,
			Validate: survey.Required,
		},
		{
			Name:     "lifetimewarning",
			Prompt:   lifetimewarningInput,
			Validate: survey.Required,
		},
	}

	return questions
}
