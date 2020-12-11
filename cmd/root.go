/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/krmdv/cli/api"
	cheerCmd "github.com/krmdv/cli/cheer"
	"github.com/krmdv/cli/config"
	loginCmd "github.com/krmdv/cli/login"
	meCmd "github.com/krmdv/cli/me"
	setupCmd "github.com/krmdv/cli/setup"
)

var version = "v0.0.1"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "karma",
	Short:   "A CLI-first dev gamification engine.",
	Version: version,
}

// Execute executes the root command.
func Execute() {
	rootCmd.Execute()
}

func init() {
	conf := config.Get()
	client := api.NewClient(config.Host(), conf.Token, conf.Team.ID, version)

	go checkVersion()

	rootCmd.AddCommand(cheerCmd.NewCmdCheer(client, conf))
	rootCmd.AddCommand(meCmd.NewCmdMe(client, conf))
	rootCmd.AddCommand(loginCmd.NewCmdLogin())
	rootCmd.AddCommand(setupCmd.NewCmdSetup(client))
}

func checkVersion() {

	// var latestRelease struct {
	// 	Name string `json:"name"`
	// }

	// gorequest.New().Get("https://api.github.com/repos/krmdv/cli/releases/latest").EndStruct(&latestRelease)

	// if version != latestRelease.Name {
	// 	color.Yellow(fmt.Sprintf("Heads up! You're running version %s of the CLI, but latest version is %s. Please upgrade at https://github.com/krmdv/cli/releases.", version, latestRelease.Name))
	// }
}
