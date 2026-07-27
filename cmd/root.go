// Package cmd implements gothub's Cobra command tree. Every command
// here must only ever read from the GitHub API — write operations are
// not implemented anywhere in this codebase, on purpose.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/longkey1/gothub/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gothub",
	Short: "Read-only GitHub CLI",
	Long: `gothub is a read-only GitHub CLI built on go-gh.

It only ever issues GET requests against the GitHub API: no command in
this codebase can create, modify, or delete anything.`,
}

func init() {
	rootCmd.Version = version.Short()
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// splitRepo parses an "OWNER/REPO" argument.
func splitRepo(spec string) (owner, repo string, err error) {
	parts := strings.SplitN(spec, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format %q, expected OWNER/REPO", spec)
	}
	return parts[0], parts[1], nil
}

// requireRepoFlag validates that --repo was provided and parses it into
// owner/repo. gothub does not auto-detect the repository from the current
// directory the way gh does, so --repo is mandatory here.
func requireRepoFlag(spec string) (owner, repo string, err error) {
	if spec == "" {
		return "", "", fmt.Errorf(`required flag(s) "repo" not set`)
	}
	return splitRepo(spec)
}

// printJSON writes v to cmd's stdout as indented JSON.
func printJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
