package cmd

import (
	"fmt"
	"strconv"
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
	Body        string    `json:"body"`
	User        user      `json:"user"`
	HTMLURL     string    `json:"html_url"`
	CreatedAt   time.Time `json:"created_at"`
	PullRequest *struct{} `json:"pull_request,omitempty"`
}

type issueComment struct {
	Body      string    `json:"body"`
	User      user      `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

type issueWithComments struct {
	issue
	Comments []issueComment `json:"comments"`
}

var (
	issueListJSON     bool
	issueListLimit    int
	issueViewJSON     bool
	issueViewComments bool
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

var issueViewCmd = &cobra.Command{
	Use:   "view <owner>/<repo> <number>",
	Short: "View an issue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := splitRepo(args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number %q", args[1])
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		var it issue
		if err := c.Get(fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, number), &it); err != nil {
			return err
		}

		var comments []issueComment
		if issueViewComments {
			comments, err = client.GetPaginated[issueComment](c, fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, number), 0)
			if err != nil {
				return err
			}
		}

		if issueViewJSON {
			if issueViewComments {
				return printJSON(cmd, issueWithComments{issue: it, Comments: comments})
			}
			return printJSON(cmd, it)
		}

		if err := printIssue(cmd, it); err != nil {
			return err
		}
		if issueViewComments {
			return printIssueComments(cmd, comments)
		}
		return nil
	},
}

func printIssue(cmd *cobra.Command, it issue) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "number:\t#%d\n", it.Number)
	_, _ = fmt.Fprintf(w, "title:\t%s\n", it.Title)
	_, _ = fmt.Fprintf(w, "state:\t%s\n", it.State)
	_, _ = fmt.Fprintf(w, "author:\t%s\n", it.User.Login)
	_, _ = fmt.Fprintf(w, "url:\t%s\n", it.HTMLURL)
	_, _ = fmt.Fprintf(w, "created at:\t%s\n", it.CreatedAt.Format(time.RFC3339))
	if err := w.Flush(); err != nil {
		return err
	}
	if it.Body != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", it.Body)
	}
	return nil
}

func printIssueComments(cmd *cobra.Command, comments []issueComment) error {
	out := cmd.OutOrStdout()
	for _, c := range comments {
		_, _ = fmt.Fprintf(out, "\n---\n%s (%s):\n%s\n", c.User.Login, c.CreatedAt.Format(time.RFC3339), c.Body)
	}
	return nil
}

func init() {
	issueListCmd.Flags().BoolVar(&issueListJSON, "json", false, "output as JSON")
	issueListCmd.Flags().IntVar(&issueListLimit, "limit", 30, "maximum number of issues to fetch")
	issueViewCmd.Flags().BoolVar(&issueViewJSON, "json", false, "output as JSON")
	issueViewCmd.Flags().BoolVar(&issueViewComments, "comments", false, "include comments")
	issueCmd.AddCommand(issueListCmd)
	issueCmd.AddCommand(issueViewCmd)
	rootCmd.AddCommand(issueCmd)
}
