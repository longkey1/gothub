package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/longkey1/gothub/internal/client"
	"github.com/spf13/cobra"
)

type searchResult[T any] struct {
	TotalCount int `json:"total_count"`
	Items      []T `json:"items"`
}

type codeResult struct {
	Path       string     `json:"path"`
	HTMLURL    string     `json:"html_url"`
	Repository repository `json:"repository"`
}

var (
	searchReposJSON   bool
	searchReposLimit  int
	searchIssuesJSON  bool
	searchIssuesLimit int
	searchPRsJSON     bool
	searchPRsLimit    int
	searchCodeJSON    bool
	searchCodeLimit   int
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search GitHub",
}

var searchReposCmd = &cobra.Command{
	Use:   "repos <query>...",
	Short: "Search repositories",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := runSearch[repository]("search/repositories", strings.Join(args, " "), searchReposLimit)
		if err != nil {
			return err
		}
		if searchReposJSON {
			return printJSON(cmd, items)
		}
		return printRepos(cmd, items)
	},
}

var searchIssuesCmd = &cobra.Command{
	Use:   "issues <query>...",
	Short: "Search issues",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := "is:issue " + strings.Join(args, " ")
		items, err := runSearch[issue]("search/issues", query, searchIssuesLimit)
		if err != nil {
			return err
		}
		if searchIssuesJSON {
			return printJSON(cmd, items)
		}
		return printIssues(cmd, items)
	},
}

var searchPRsCmd = &cobra.Command{
	Use:   "prs <query>...",
	Short: "Search pull requests",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := "is:pr " + strings.Join(args, " ")
		items, err := runSearch[issue]("search/issues", query, searchPRsLimit)
		if err != nil {
			return err
		}
		if searchPRsJSON {
			return printJSON(cmd, items)
		}
		return printIssues(cmd, items)
	},
}

var searchCodeCmd = &cobra.Command{
	Use:   "code <query>...",
	Short: "Search code",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := runSearch[codeResult]("search/code", strings.Join(args, " "), searchCodeLimit)
		if err != nil {
			return err
		}
		if searchCodeJSON {
			return printJSON(cmd, items)
		}
		return printCodeResults(cmd, items)
	},
}

// runSearch issues a single request against a GitHub Search API endpoint
// and returns up to limit items (limit<=0 defaults to 30, capped at 100
// since Search API responses are not paginated via GetPaginated).
func runSearch[T any](path, query string, limit int) ([]T, error) {
	c, err := client.NewReadOnlyClient()
	if err != nil {
		return nil, err
	}

	perPage := limit
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}

	var result searchResult[T]
	requestPath := fmt.Sprintf("%s?q=%s&per_page=%d", path, url.QueryEscape(query), perPage)
	if err := c.Get(requestPath, &result); err != nil {
		return nil, err
	}

	items := result.Items
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func printCodeResults(cmd *cobra.Command, items []codeResult) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "REPOSITORY\tPATH\tURL")
	for _, it := range items {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", it.Repository.FullName, it.Path, it.HTMLURL)
	}
	return w.Flush()
}

func init() {
	searchReposCmd.Flags().BoolVar(&searchReposJSON, "json", false, "output as JSON")
	searchReposCmd.Flags().IntVar(&searchReposLimit, "limit", 30, "maximum number of repositories to fetch")

	searchIssuesCmd.Flags().BoolVar(&searchIssuesJSON, "json", false, "output as JSON")
	searchIssuesCmd.Flags().IntVar(&searchIssuesLimit, "limit", 30, "maximum number of issues to fetch")

	searchPRsCmd.Flags().BoolVar(&searchPRsJSON, "json", false, "output as JSON")
	searchPRsCmd.Flags().IntVar(&searchPRsLimit, "limit", 30, "maximum number of pull requests to fetch")

	searchCodeCmd.Flags().BoolVar(&searchCodeJSON, "json", false, "output as JSON")
	searchCodeCmd.Flags().IntVar(&searchCodeLimit, "limit", 30, "maximum number of code results to fetch")

	searchCmd.AddCommand(searchReposCmd)
	searchCmd.AddCommand(searchIssuesCmd)
	searchCmd.AddCommand(searchPRsCmd)
	searchCmd.AddCommand(searchCodeCmd)
	rootCmd.AddCommand(searchCmd)
}
