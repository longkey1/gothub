package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/longkey1/gothub/internal/client"
	"github.com/spf13/cobra"
)

type repository struct {
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	HTMLURL         string `json:"html_url"`
	DefaultBranch   string `json:"default_branch"`
	Language        string `json:"language"`
	Private         bool   `json:"private"`
	Fork            bool   `json:"fork"`
	Archived        bool   `json:"archived"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	OpenIssuesCount int    `json:"open_issues_count"`
}

var repoViewJSON bool

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "View repository information",
}

var repoViewCmd = &cobra.Command{
	Use:   "view <owner>/<repo>",
	Short: "View a repository",
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

		var r repository
		if err := c.Get(fmt.Sprintf("repos/%s/%s", owner, repo), &r); err != nil {
			return err
		}

		if repoViewJSON {
			return printJSON(cmd, r)
		}
		return printRepo(cmd, r)
	},
}

func printRepo(cmd *cobra.Command, r repository) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "name:\t%s\n", r.FullName)
	_, _ = fmt.Fprintf(w, "description:\t%s\n", r.Description)
	_, _ = fmt.Fprintf(w, "url:\t%s\n", r.HTMLURL)
	_, _ = fmt.Fprintf(w, "default branch:\t%s\n", r.DefaultBranch)
	_, _ = fmt.Fprintf(w, "language:\t%s\n", r.Language)
	_, _ = fmt.Fprintf(w, "private:\t%t\n", r.Private)
	_, _ = fmt.Fprintf(w, "fork:\t%t\n", r.Fork)
	_, _ = fmt.Fprintf(w, "archived:\t%t\n", r.Archived)
	_, _ = fmt.Fprintf(w, "stars:\t%d\n", r.StargazersCount)
	_, _ = fmt.Fprintf(w, "forks:\t%d\n", r.ForksCount)
	_, _ = fmt.Fprintf(w, "open issues:\t%d\n", r.OpenIssuesCount)
	return w.Flush()
}

func init() {
	repoViewCmd.Flags().BoolVar(&repoViewJSON, "json", false, "output as JSON")
	repoCmd.AddCommand(repoViewCmd)
	rootCmd.AddCommand(repoCmd)
}
