package config

import (
	"log"
	"os"
	"path"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config stores the vmpooler endpoint and user token
type Config struct {
	Endpoint        string `yaml:"endpoint"`
	Token           string `yaml:"token"`
	LifetimeWarning int    `yaml:"lifetimewarning"`
}

// DefaultDir is the default vmpooler-bitbar config directory
const DefaultDir string = "~/.vmpooler-bitbar"
const yamlFile = "config.yaml"

var DefaultConfig = Config{
	Endpoint:        "https://vmpooler.mycompany.net/api/v1",
	Token:           "myvmpoolertoken",
	LifetimeWarning: 1,
}

// Dir returns path to the config dir
func Dir() string {
	cfgPath, _ := homedir.Expand(DefaultDir)
	return path.Clean(cfgPath)
}

// File returns path to the config file
func File() string {
	return path.Clean(filepath.Join(Dir(), yamlFile))
}

// EnsureConfigDir creates a configDir() if it doesn't already exist
func EnsureConfigDir() error {
	dir := Dir()
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return nil
	}
	err := os.Mkdir(dir, 0700)
	if err != nil {
		return err
	}
	return nil
}

// Read config from the specified dir
func Read() (Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(Dir())

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	conf := new(Config)

	err = mapstructure.Decode(viper.AllSettings(), conf)
	if err != nil {
		log.Fatalf("%v", err)
	}

	return *conf, nil
}
