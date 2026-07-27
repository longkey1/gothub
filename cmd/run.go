package cmd

import (
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/longkey1/gothub/internal/client"
	"github.com/spf13/cobra"
)

type workflowRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
}

type workflowJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
}

type workflowJobsResponse struct {
	Jobs []workflowJob `json:"jobs"`
}

type workflowRunWithJobs struct {
	workflowRun
	Jobs []workflowJob `json:"jobs"`
}

type workflowRunsResponse struct {
	WorkflowRuns []workflowRun `json:"workflow_runs"`
}

var (
	runRepo          string
	runViewJSON      bool
	runViewLog       bool
	runViewLogFailed bool
	runListJSON      bool
	runListLimit     int
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "View workflow runs",
}

var runListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflow runs",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := requireRepoFlag(runRepo)
		if err != nil {
			return err
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		perPage := runListLimit
		if perPage <= 0 || perPage > 100 {
			perPage = 30
		}

		var resp workflowRunsResponse
		path := fmt.Sprintf("repos/%s/%s/actions/runs?per_page=%d", owner, repo, perPage)
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		items := resp.WorkflowRuns
		if runListLimit > 0 && len(items) > runListLimit {
			items = items[:runListLimit]
		}

		if runListJSON {
			return printJSON(cmd, items)
		}
		return printRuns(cmd, items)
	},
}

func printRuns(cmd *cobra.Command, items []workflowRun) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tBRANCH\tSTATUS\tCONCLUSION")
	for _, r := range items {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", r.ID, r.Name, r.HeadBranch, r.Status, r.Conclusion)
	}
	return w.Flush()
}

var runViewCmd = &cobra.Command{
	Use:   "view <run-id>",
	Short: "View a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := requireRepoFlag(runRepo)
		if err != nil {
			return err
		}
		runID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid run id %q", args[0])
		}

		c, err := client.NewReadOnlyClient()
		if err != nil {
			return err
		}

		var run workflowRun
		if err := c.Get(fmt.Sprintf("repos/%s/%s/actions/runs/%d", owner, repo, runID), &run); err != nil {
			return err
		}

		var jobsResp workflowJobsResponse
		if err := c.Get(fmt.Sprintf("repos/%s/%s/actions/runs/%d/jobs", owner, repo, runID), &jobsResp); err != nil {
			return err
		}

		if runViewLog || runViewLogFailed {
			return printRunLogs(cmd, c, owner, repo, jobsResp.Jobs, runViewLogFailed)
		}

		if runViewJSON {
			return printJSON(cmd, workflowRunWithJobs{workflowRun: run, Jobs: jobsResp.Jobs})
		}
		return printRun(cmd, run, jobsResp.Jobs)
	},
}

func printRun(cmd *cobra.Command, run workflowRun, jobs []workflowJob) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "name:\t%s\n", run.Name)
	_, _ = fmt.Fprintf(w, "status:\t%s\n", run.Status)
	_, _ = fmt.Fprintf(w, "conclusion:\t%s\n", run.Conclusion)
	_, _ = fmt.Fprintf(w, "branch:\t%s\n", run.HeadBranch)
	_, _ = fmt.Fprintf(w, "url:\t%s\n", run.HTMLURL)
	if err := w.Flush(); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	jw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(jw, "JOB\tSTATUS\tCONCLUSION")
	for _, j := range jobs {
		_, _ = fmt.Fprintf(jw, "%s\t%s\t%s\n", j.Name, j.Status, j.Conclusion)
	}
	return jw.Flush()
}

func printRunLogs(cmd *cobra.Command, c *client.ReadOnlyClient, owner, repo string, jobs []workflowJob, failedOnly bool) error {
	out := cmd.OutOrStdout()
	for _, j := range jobs {
		if failedOnly && j.Conclusion != "failure" {
			continue
		}
		data, err := c.GetRaw(fmt.Sprintf("repos/%s/%s/actions/jobs/%d/logs", owner, repo, j.ID))
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "\n=== %s ===\n%s\n", j.Name, data)
	}
	return nil
}

func init() {
	runCmd.PersistentFlags().StringVarP(&runRepo, "repo", "R", "", "Select another repository using the OWNER/REPO format")
	runViewCmd.Flags().BoolVar(&runViewJSON, "json", false, "output as JSON")
	runViewCmd.Flags().BoolVar(&runViewLog, "log", false, "show logs for all jobs")
	runViewCmd.Flags().BoolVar(&runViewLogFailed, "log-failed", false, "show logs for failed jobs only")
	runListCmd.Flags().BoolVar(&runListJSON, "json", false, "output as JSON")
	runListCmd.Flags().IntVar(&runListLimit, "limit", 30, "maximum number of workflow runs to fetch")
	runCmd.AddCommand(runViewCmd)
	runCmd.AddCommand(runListCmd)
	rootCmd.AddCommand(runCmd)
}
