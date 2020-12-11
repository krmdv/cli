package config

import (
	"errors"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Configuration is the
type Configuration struct {
	Token string `mapstructure:"token"`
	Team  struct {
		ID string `mapstructure:"id"`
	} `mapstructure:"team"`
	Users []struct {
		Name string `mapstructure:"name"`
		ID   string `mapstructure:"id"`
	} `mapstructure:"users"`
	Feats []struct {
		ID    string `mapstructure:"id"`
		Label string `mapstructure:"label"`
		Slug  string `mapstructure:"slug"`
		Karma int    `mapstructure:"karma"`
	} `mapstructure:"feats"`
}

// CheckAuthed ensures user has setup an API token
func CheckAuthed() error {
	if token := viper.GetString("token"); token == "" {
		return errors.New("no token present, please run 'karma config --token xxx' first")
	}

	return nil
}

// CheckLoaded ensures configuration has been loaded
func CheckLoaded() error {
	if token := viper.GetString("token"); token == "" {
		return errors.New("no token present, please run 'karma config --token xxx' first")
	}

	if teamID := viper.GetString("team.id"); teamID == "" {
		return errors.New("no github org present, please run 'karma config --org xxx' first")
	}

	return nil
}

// Host returns the base Karma API endpoint
func Host() string {
	host := os.Getenv("KARMA_HOST")

	if host == "" {
		host = "https://api.getkarma.dev"
	}

	return host
}

// Get returns a configuration object
func Get() Configuration {
	home, err := homedir.Dir()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".karma")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no issue, we'll warn users about missing conf when running commands
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	var conf Configuration
	viper.Unmarshal(&conf)

	return conf
}
