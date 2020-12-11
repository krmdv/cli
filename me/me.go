package me

import (
	"fmt"
	"strconv"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/krmdv/cli/api"
	"github.com/krmdv/cli/config"
	"github.com/spf13/cobra"
)

type dashboard struct {
	User struct {
		ID    string `json:"id"`
		Stats struct {
			Level            int   `json:"level"`
			Multiplier       int   `json:"multiplier"`
			Progress         int   `json:"progress"`
			KarmaToNextLevel int64 `json:"karmaToNextLevel"`
		} `json:"stats"`
		TotalAccruedKarma int64 `json:"totalAccruedKarma"`
	} `json:"user"`
	Leaderboard struct {
		Names  []string  `json:"names"`
		Levels []float64 `json:"levels"`
	} `json:"leaderboard"`
	Logs []struct {
		Ago         string `json:"ago"`
		From        string `json:"from"`
		ToUser      string `json:"toUser"`
		Karma       int    `json:"karma"`
		FeatID      string `json:"featId"`
		CurrentUser bool   `json:"currentUser"`
	} `json:"logs"`
}

// NewCmdMe displays dashboard
func NewCmdMe(client api.Client, conf config.Configuration) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "me",
		Short: "Display your Karma dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.CheckLoaded(); err != nil {
				return err
			}

			return meRun(client, conf)
		},
	}

	return cmd
}

func meRun(client api.Client, conf config.Configuration) error {
	var data dashboard

	client.Get("/dashboard", &data)

	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	welcomeBox := textBox("👋 Welcome", fmt.Sprintf("You are a level %v karma developer.\nYou earned %s karma pts in total.\nHit Q to quit.", data.User.Stats.Level, formatNumber(data.User.TotalAccruedKarma)), true)
	gauge := gauge()
	multiplierBox := textBox("Mul.", fmt.Sprintf(" %vx", data.User.Stats.Multiplier), false)
	featsData := make([]string, 0, len(conf.Feats))
	featsList := list("Feats (SLUG)", featsData)
	tip := textBox("Tip", "Cheer with 'karma c DEV_NAME -f SLUG'", true)
	footer := textBox("Info", "Running karma CLI v1.0.0. Check docs at https://docs.getkarma.dev. Thanks for being awesome 😍.", true)
	leaderBoard := leaderboard(data.Leaderboard.Names, data.Leaderboard.Levels)
	logs := logs(data, conf)

	welcomeBox.SetRect(0, 0, 42, 5)
	gauge.SetRect(0, 5, 34, 8)
	multiplierBox.SetRect(35, 5, 42, 8)
	featsList.SetRect(0, 8, 42, 24)
	tip.SetRect(0, 24, 42, 27)
	footer.SetRect(0, 27, 100, 30)
	leaderBoard.SetRect(43, 0, 100, 8)
	logs.SetRect(43, 8, 100, 27)

	for _, f := range conf.Feats {
		featsData = append(featsData, fmt.Sprintf("%s (%s)", f.Label, f.Slug))
	}

	gaugePercent := data.User.Stats.Progress

	draw := func(count int) {
		if gauge.Percent < gaugePercent {
			gauge.Percent++
			gauge.Label = fmt.Sprintf("%v%%", gauge.Percent)
		} else {
			gauge.Label = fmt.Sprintf("%v%% - %s pts to lvl %v", gaugePercent, formatNumber(data.User.Stats.KarmaToNextLevel), data.User.Stats.Level+1)
		}

		if count%20 == 0 {
			featsList.Rows = featsData[(count/20)%(len(featsData)):]
		}

		ui.Render(welcomeBox, multiplierBox, gauge, featsList, tip, footer, leaderBoard, logs)
	}

	tickerCount := 1
	draw(tickerCount)
	tickerCount++
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second / 20).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			}
		case <-ticker:
			draw(tickerCount)
			tickerCount++
		}
	}

}

func textBox(title string, text string, grey bool) *widgets.Paragraph {
	w := widgets.NewParagraph()
	w.Title = title
	w.Text = text
	w.TextStyle.Fg = ui.ColorBlue
	w.TextStyle.Modifier = ui.ModifierBold

	if grey {
		w.TextStyle.Fg = ui.Color(245)
		w.TextStyle.Modifier = ui.ModifierClear
	}

	w.BorderStyle.Fg = ui.ColorYellow

	return w
}

func gauge() *widgets.Gauge {
	w := widgets.NewGauge()
	w.Title = "Karma pts"
	w.Percent = 0
	w.BarColor = ui.ColorBlue
	w.BorderStyle.Fg = ui.ColorYellow
	w.TitleStyle.Fg = ui.ColorWhite

	return w
}

func leaderboard(labels []string, values []float64) *widgets.BarChart {
	w := widgets.NewBarChart()
	w.Title = "Team leaderboard"
	w.BorderStyle.Fg = ui.ColorYellow
	w.BarColors = []ui.Color{ui.ColorBlue}
	w.BarGap = 2
	w.PaddingLeft = 2
	w.PaddingTop = 0
	w.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	w.NumStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	w.Labels = labels
	w.Data = values

	return w
}

func list(title string, data []string) *widgets.List {
	w := widgets.NewList()
	w.Title = title
	w.Rows = data
	w.TextStyle.Fg = ui.Color(245)
	w.BorderStyle.Fg = ui.ColorYellow

	return w
}

func logs(data dashboard, conf config.Configuration) *widgets.Table {
	w := widgets.NewTable()

	w.RowSeparator = false
	w.BorderStyle.Fg = ui.ColorYellow
	w.TextStyle = ui.NewStyle(ui.Color(245))
	w.ColumnWidths = []int{8, 13, 13, 12, 8}
	w.RowStyles[0] = ui.NewStyle(ui.ColorYellow, ui.ColorBlack, ui.ModifierBold)

	w.Title = "Karma events log"

	w.Rows = [][]string{
		{"Ago", "From", "To", "Feat", "Karma"},
	}

	for k, v := range data.Logs {

		var feat string
		for _, f := range conf.Feats {
			if f.ID == v.FeatID {
				feat = f.Label
				break
			}
		}

		w.Rows = append(w.Rows, []string{v.Ago, v.From, v.ToUser, feat, fmt.Sprintf("%v", v.Karma)})

		modifier := ui.ModifierClear
		color := ui.Color(245)

		if v.CurrentUser {
			modifier = ui.ModifierBold
			if v.Karma > 0 {
				color = ui.ColorGreen
			} else {
				color = ui.ColorRed
			}
		}

		w.RowStyles[k+1] = ui.NewStyle(color, ui.ColorClear, modifier)
	}

	return w
}

func formatNumber(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}
