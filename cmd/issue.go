package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/longkey1/gothub/internal/client"
	"github.com/spf13/cobra"
)

type user struct {
	Login string `json:"login"`
}

type issue struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	User        user      `json:"user"`
	HTMLURL     string    `json:"html_url"`
	CreatedAt   time.Time `json:"created_at"`
	PullRequest *struct{} `json:"pull_request,omitempty"`
}

var (
	issueListJSON  bool
	issueListLimit int
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "View issues",
}

var issueListCmd = &cobra.Command{
	Use:   "list <owner>/<repo>",
	Short: "List issues",
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

		// GitHub's issues endpoint also returns pull requests; filter
		// them out so "issue list" only shows actual issues.
		items, err := client.GetPaginated[issue](c, fmt.Sprintf("repos/%s/%s/issues", owner, repo), 0)
		if err != nil {
			return err
		}
		issuesOnly := make([]issue, 0, len(items))
		for _, it := range items {
			if it.PullRequest == nil {
				issuesOnly = append(issuesOnly, it)
			}
			if issueListLimit > 0 && len(issuesOnly) >= issueListLimit {
				break
			}
		}

		if issueListJSON {
			return printJSON(cmd, issuesOnly)
		}
		return printIssues(cmd, issuesOnly)
	},
}

func printIssues(cmd *cobra.Command, items []issue) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NUMBER\tTITLE\tSTATE\tAUTHOR")
	for _, it := range items {
		_, _ = fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n", it.Number, it.Title, it.State, it.User.Login)
	}
	return w.Flush()
}

func init() {
	issueListCmd.Flags().BoolVar(&issueListJSON, "json", false, "output as JSON")
	issueListCmd.Flags().IntVar(&issueListLimit, "limit", 30, "maximum number of issues to fetch")
	issueCmd.AddCommand(issueListCmd)
	rootCmd.AddCommand(issueCmd)
}
