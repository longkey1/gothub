package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/longkey1/gothub/internal/client"
	"github.com/spf13/cobra"
)

type pullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	User      user      `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	prListJSON  bool
	prListLimit int
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "View pull requests",
}

var prListCmd = &cobra.Command{
	Use:   "list <owner>/<repo>",
	Short: "List pull requests",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := splitRepo(args[0])
		if err != nil {
			return err
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		items, err := client.GetPaginated[pullRequest](c, fmt.Sprintf("repos/%s/%s/pulls", owner, repo), prListLimit)
		if err != nil {
			return err
		}

		if prListJSON {
			return printJSON(cmd, items)
		}
		return printPullRequests(cmd, items)
	},
}

func printPullRequests(cmd *cobra.Command, items []pullRequest) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NUMBER\tTITLE\tSTATE\tAUTHOR")
	for _, it := range items {
		_, _ = fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n", it.Number, it.Title, it.State, it.User.Login)
	}
	return w.Flush()
}

func init() {
	prListCmd.Flags().BoolVar(&prListJSON, "json", false, "output as JSON")
	prListCmd.Flags().IntVar(&prListLimit, "limit", 30, "maximum number of pull requests to fetch")
	prCmd.AddCommand(prListCmd)
	rootCmd.AddCommand(prCmd)
}
