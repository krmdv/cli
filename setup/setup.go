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

package setup

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/krmdv/cli/api"
	"github.com/krmdv/cli/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdSetup creates a cheer command
func NewCmdSetup(client api.Client) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Configure Karma",
		Long:  `Configure Karma`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.CheckAuthed(); err != nil {
				return err
			}

			org, _ := cmd.Flags().GetString("org")
			slackWebhookURL, _ := cmd.Flags().GetString("slack")
			printGithub, _ := cmd.Flags().GetBool("github")
			printSentry, _ := cmd.Flags().GetBool("sentry")
			return setupRun(client, org, slackWebhookURL, printGithub, printSentry)
		},
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringP("org", "o", "", "set active github organization")
	cmd.Flags().StringP("slack", "s", "", "set your slack notifications webhook URL")
	cmd.Flags().Bool("github", false, "print your Github webhook URL")
	cmd.Flags().Bool("sentry", false, "prints your Sentry webhook URL")

	return cmd
}

func setupRun(client api.Client, org string, slackWebhookURL string, printGithub bool, printSentry bool) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)

	s.Start()

	if org != "" {
		type setTeamPayload struct {
			GithubLogin string `json:"githubLogin"`
		}

		type setTeamResp struct {
			ID    string `json:"id"`
			Token string `json:"apiToken"`
			Name  string `json:"name"`
			Users []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"users"`
		}

		type getFeatsResp []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Karma int    `json:"karma"`
			Slug  string `json:"slug"`
		}

		var teamResp setTeamResp
		var featsResp getFeatsResp

		// Get team configuration (id and members)
		if err := client.Post("/teams", setTeamPayload{GithubLogin: org}, &teamResp); err != nil {
			return err
		}

		if err := client.Get("/feats", &featsResp); err != nil {
			return err
		}

		viper.Set("team.id", teamResp.ID)
		viper.Set("team.token", teamResp.Token)
		viper.Set("team.name", teamResp.Name)
		viper.Set("users", teamResp.Users)
		viper.Set("feats", featsResp)

		viper.WriteConfig()
	}

	if slackWebhookURL != "" {
		type setSlackWebhookURLPayload struct {
			SlackWebhookURL string `json:"slackWebhookUrl"`
		}

		if err := client.Post("/teams/current/slack-webhook-url", setSlackWebhookURLPayload{SlackWebhookURL: slackWebhookURL}, `{}`); err != nil {
			return err
		}
	}

	if printGithub {
		fmt.Print("👉 Navigate to the following link: ")
		color.Yellow("https://github.com/organizations/%s/settings/hooks/new", viper.GetString("team.name"))
		fmt.Println()
		fmt.Print("* Set the 'Payload URL' field to: ")
		color.Blue("https://api.getkarma.dev/events/github?token=" + viper.GetString("team.token"))
		fmt.Print("* Set 'Content type' field to: ")
		color.Blue("application/json")
		fmt.Println("* Leave 'Secret' field blank.")
		fmt.Print("* Select the following 'individual events': ")
		color.Blue("Pull request reviews, Statuses, Pull request review comments, Pull requests")
		fmt.Println()
		fmt.Println("Then save changes.")
	}

	if printSentry {
		fmt.Print("👉 Navigate to the following link: ")
		color.Yellow("https://sentry.io/settings/%s/developer-settings/new-internal/", viper.GetString("team.name"))
		fmt.Println("(you might need to change this URL to reflect your Sentry org name if it differs from Github's)")
		fmt.Println()
		fmt.Print("* Set the 'Webhook URL' field to: ")
		color.Blue("https://api.getkarma.dev/events/sentry?token=" + viper.GetString("team.token"))
		fmt.Print("* Set 'Issues & Events' permission field to: ")
		color.Blue("Read")
		fmt.Print("* Select the following check in the 'Webhooks' box: ")
		color.Blue("issue (created, resolved, assigned)")
		fmt.Println()
		fmt.Println("Then save changes.")
	}

	if err := client.Post("/users/me/setup", `{}`, `{}`); err != nil {
		return err
	}

	s.Stop()

	color.Green("✅ All set! You're ready to spread good karma.")

	return nil
}
