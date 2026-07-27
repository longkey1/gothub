package cmd

import (
	"fmt"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/longkey1/gothub/internal/client"
	"github.com/spf13/cobra"
)

type pullRequestHead struct {
	SHA string `json:"sha"`
}

type pullRequest struct {
	Number    int             `json:"number"`
	Title     string          `json:"title"`
	State     string          `json:"state"`
	Body      string          `json:"body"`
	MergedAt  *time.Time      `json:"merged_at"`
	User      user            `json:"user"`
	HTMLURL   string          `json:"html_url"`
	CreatedAt time.Time       `json:"created_at"`
	Head      pullRequestHead `json:"head"`
}

// displayState returns "merged" for merged pull requests, since the
// GitHub API's own "state" field only ever holds "open" or "closed".
func (pr pullRequest) displayState() string {
	if pr.MergedAt != nil {
		return "merged"
	}
	return pr.State
}

type reviewComment struct {
	Path      string    `json:"path"`
	Body      string    `json:"body"`
	User      user      `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

type pullRequestWithComments struct {
	pullRequest
	Comments []reviewComment `json:"comments"`
}

type prFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
}

type checkRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
}

type checkRunsResponse struct {
	CheckRuns []checkRun `json:"check_runs"`
}

var (
	prRepo         string
	prListJSON     bool
	prListLimit    int
	prViewJSON     bool
	prViewComments bool
	prDiffNameOnly bool
	prDiffJSON     bool
	prChecksJSON   bool
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "View pull requests",
}

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := requireRepoFlag(prRepo)
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
		_, _ = fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n", it.Number, it.Title, it.displayState(), it.User.Login)
	}
	return w.Flush()
}

var prViewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "View a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := requireRepoFlag(prRepo)
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid pull request number %q", args[0])
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		var pr pullRequest
		if err := c.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), &pr); err != nil {
			return err
		}

		var comments []reviewComment
		if prViewComments {
			comments, err = client.GetPaginated[reviewComment](c, fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, number), 0)
			if err != nil {
				return err
			}
		}

		if prViewJSON {
			if prViewComments {
				return printJSON(cmd, pullRequestWithComments{pullRequest: pr, Comments: comments})
			}
			return printJSON(cmd, pr)
		}

		if err := printPullRequest(cmd, pr); err != nil {
			return err
		}
		if prViewComments {
			return printReviewComments(cmd, comments)
		}
		return nil
	},
}

func printPullRequest(cmd *cobra.Command, pr pullRequest) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "number:\t#%d\n", pr.Number)
	_, _ = fmt.Fprintf(w, "title:\t%s\n", pr.Title)
	_, _ = fmt.Fprintf(w, "state:\t%s\n", pr.displayState())
	_, _ = fmt.Fprintf(w, "author:\t%s\n", pr.User.Login)
	_, _ = fmt.Fprintf(w, "url:\t%s\n", pr.HTMLURL)
	_, _ = fmt.Fprintf(w, "created at:\t%s\n", pr.CreatedAt.Format(time.RFC3339))
	if err := w.Flush(); err != nil {
		return err
	}
	if pr.Body != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", pr.Body)
	}
	return nil
}

func printReviewComments(cmd *cobra.Command, comments []reviewComment) error {
	out := cmd.OutOrStdout()
	for _, c := range comments {
		_, _ = fmt.Fprintf(out, "\n---\n%s @ %s (%s):\n%s\n", c.User.Login, c.Path, c.CreatedAt.Format(time.RFC3339), c.Body)
	}
	return nil
}

var prDiffCmd = &cobra.Command{
	Use:   "diff <number>",
	Short: "View changes in a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := requireRepoFlag(prRepo)
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid pull request number %q", args[0])
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		files, err := client.GetPaginated[prFile](c, fmt.Sprintf("repos/%s/%s/pulls/%d/files", owner, repo, number), 0)
		if err != nil {
			return err
		}

		if prDiffJSON {
			return printJSON(cmd, files)
		}
		return printPRFiles(cmd, files)
	},
}

func printPRFiles(cmd *cobra.Command, files []prFile) error {
	out := cmd.OutOrStdout()
	if prDiffNameOnly {
		for _, f := range files {
			_, _ = fmt.Fprintln(out, f.Filename)
		}
		return nil
	}
	for _, f := range files {
		_, _ = fmt.Fprintf(out, "diff --git a/%s b/%s\n", f.Filename, f.Filename)
		_, _ = fmt.Fprintf(out, "--- a/%s\n+++ b/%s\n", f.Filename, f.Filename)
		_, _ = fmt.Fprintln(out, f.Patch)
	}
	return nil
}

var prChecksCmd = &cobra.Command{
	Use:   "checks <number>",
	Short: "View CI check status for a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := requireRepoFlag(prRepo)
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid pull request number %q", args[0])
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		var pr pullRequest
		if err := c.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), &pr); err != nil {
			return err
		}

		var checks checkRunsResponse
		if err := c.Get(fmt.Sprintf("repos/%s/%s/commits/%s/check-runs", owner, repo, pr.Head.SHA), &checks); err != nil {
			return err
		}

		if prChecksJSON {
			return printJSON(cmd, checks.CheckRuns)
		}
		return printChecks(cmd, checks.CheckRuns)
	},
}

func printChecks(cmd *cobra.Command, runs []checkRun) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tSTATUS\tCONCLUSION")
	for _, r := range runs {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.Status, r.Conclusion)
	}
	return w.Flush()
}

func init() {
	prCmd.PersistentFlags().StringVarP(&prRepo, "repo", "R", "", "Select another repository using the OWNER/REPO format")
	prListCmd.Flags().BoolVar(&prListJSON, "json", false, "output as JSON")
	prListCmd.Flags().IntVar(&prListLimit, "limit", 30, "maximum number of pull requests to fetch")
	prViewCmd.Flags().BoolVar(&prViewJSON, "json", false, "output as JSON")
	prViewCmd.Flags().BoolVar(&prViewComments, "comments", false, "include review comments")
	prDiffCmd.Flags().BoolVar(&prDiffNameOnly, "name-only", false, "show only names of changed files")
	prDiffCmd.Flags().BoolVar(&prDiffJSON, "json", false, "output as JSON")
	prChecksCmd.Flags().BoolVar(&prChecksJSON, "json", false, "output as JSON")
	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prViewCmd)
	prCmd.AddCommand(prDiffCmd)
	prCmd.AddCommand(prChecksCmd)
	rootCmd.AddCommand(prCmd)
}
