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

package cheer

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"github.com/krmdv/cli/api"
	"github.com/krmdv/cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Users is
type Users []struct {
	Name string `mapstructure:"name"`
	ID   string `mapstructure:"id"`
}

// Feats is
type Feats []struct {
	ID    string `mapstructure:"id"`
	Label string `mapstructure:"label"`
	Slug  string `mapstructure:"slug"`
	Karma int    `mapstructure:"karma"`
}

// NewCmdCheer creates a cheer command
func NewCmdCheer(client api.Client, conf config.Configuration) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "cheer <developer>",
		Short: "Cheer a dev",
		Long: heredoc.Doc(`
			Cheer a developer for a given feat.

			The available developers and feats are those configured for your current team.
			
			Note that you cannot cheer yourself.
		`),
		Example: heredoc.Doc(`
			# start cheering from self-serve menu
			$ karma c

			# cheer John Doe and pick feat manually
			$ karma c johndoe

			# cheer Dab Abramov for being a React Guru
			$ karma c gaearon -f react -msg "Well done, Dan!"

			# cheer Troy Hunt for being a Super Hacker
			$ karma c troyhunt -f hacker -msg "Nothing like DDoS for breakfast!"
		`),
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Get()

			if err := config.CheckLoaded(); err != nil {
				return err
			}

			return cheerRun(client, conf, args, cmd.Flags())
		},
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringP("feat", "f", "", "The slug of the feat to cheer the dev for")
	cmd.Flags().StringP("msg", "m", "", "An optional message for this dev")

	return cmd
}

func cheerRun(client api.Client, conf config.Configuration, args []string, flags *pflag.FlagSet) error {
	len := len(args)

	user := ""
	feat, _ := flags.GetString("feat")
	msg, _ := flags.GetString("msg")

	if len >= 1 {
		user = args[0]
	}

	askForMsg := feat == "" && msg == ""

	// Find user from argument or from prompt
	userID := ""
	var users []string
	for _, m := range conf.Users {
		if m.Name == user {
			userID = m.ID
			break
		} else {
			users = append(users, m.Name)
		}
	}

	if userID == "" {
		err := survey.AskOne(&survey.Select{
			Message: "Who do you want to cheer?",
			Options: users,
		}, &user, survey.WithValidator(survey.Required))

		if err == terminal.InterruptErr {
			fmt.Println("interrupted")

			os.Exit(0)
		} else if err != nil {
			panic(err)
		}

		for _, d := range conf.Users {
			if d.Name == user && d.ID != "" {
				userID = d.ID
				break
			}
		}
	}

	// Find feat from argument or from prompt
	featID := ""
	var feats []string
	for _, f := range conf.Feats {
		if f.Karma > 0 {
			if f.Slug == feat {
				featID = f.ID
				break
			} else {
				feats = append(feats, f.Label)
			}
		}
	}

	if featID == "" {
		err := survey.AskOne(&survey.Select{
			Message: "What to cheer that dev for?",
			Options: feats,
		}, &feat, survey.WithValidator(survey.Required))

		if err == terminal.InterruptErr {
			fmt.Println("interrupted")

			os.Exit(0)
		} else if err != nil {
			panic(err)
		}

		for _, f := range conf.Feats {
			if f.Label == feat {
				featID = f.ID
				break
			}
		}
	}

	// Find message from argument or from prompt
	if msg == "" && askForMsg {
		err := survey.AskOne(&survey.Input{
			Message: "Any comment?",
		}, &msg, survey.WithValidator(survey.MaxLength(42)))

		if err == terminal.InterruptErr {
			fmt.Println("interrupted")

			os.Exit(0)
		} else if err != nil {
			panic(err)
		}
	}

	type cheerPayload struct {
		UserID string `json:"toUserId"`
		FeatID string `json:"featId"`
	}

	type cheerRes struct {
		DeliveredToActiveUser bool `json:"deliveredToActiveUser"`
		Karma                 int  `json:"karma"`
	}

	var res cheerRes

	if err := client.Post("/cheers", cheerPayload{
		UserID: userID,
		FeatID: featID,
	}, &res); err != nil {
		return err
	}

	if !res.DeliveredToActiveUser {
		color.Yellow(fmt.Sprintf("Uh-oh 🤭: %s has received your cheer but has no active Karma account yet - consider inviting that dev to spread the love 💌 .", user))
	}

	color.Green(fmt.Sprintf("You rock, thanks for spreading good karma! %s got %v points thanks to your cheer.", user, res.Karma))

	return nil
}
