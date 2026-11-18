package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const formatMarkdown = "markdown"

// validateFormat checks a --format value before any API request is made.
// An empty value means the default human-readable text output.
func validateFormat(format string) error {
	if format == "" || format == formatMarkdown {
		return nil
	}
	return fmt.Errorf("unsupported format %q, expected %q", format, formatMarkdown)
}

// markdownComment is one comment rendered under the "# Comments" section.
type markdownComment struct {
	Heading string
	Body    string
}

// printMarkdown writes a Markdown document with frontmatter as YAML
// frontmatter, body as-is, and comments (when withComments is set) under
// a "# Comments" section. Issue and PR bodies conventionally use "##" or
// deeper headings, so the top-level "#" marks the boundary between the
// author's body and the comments gothub appends.
func printMarkdown(cmd *cobra.Command, frontmatter any, body string, withComments bool, comments []markdownComment) error {
	fm, err := yaml.Marshal(frontmatter)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("---\n")
	writeMarkdownBody(&b, body)
	if withComments {
		b.WriteString("\n# Comments\n")
		for _, c := range comments {
			_, _ = fmt.Fprintf(&b, "\n## %s\n", c.Heading)
			writeMarkdownBody(&b, c.Body)
		}
	}

	_, err = fmt.Fprint(cmd.OutOrStdout(), b.String())
	return err
}

// writeMarkdownBody writes s preceded by a blank line, converting CRLF to
// LF since bodies written in the GitHub web UI are stored with CRLF line
// endings.
func writeMarkdownBody(b *strings.Builder, s string) {
	s = strings.TrimRight(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if s == "" {
		return
	}
	_, _ = fmt.Fprintf(b, "\n%s\n", s)
}
