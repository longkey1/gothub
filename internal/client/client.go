// Package client wraps go-gh's REST client and exposes only read
// operations, so gothub cannot perform write requests against the
// GitHub API by construction.
package client

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/cli/go-gh/v2/pkg/api"
)

var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

// ReadOnlyClient wraps api.RESTClient. Only Get and GetPaginated are
// exposed; Post/Patch/Put/Delete are intentionally never implemented.
type ReadOnlyClient struct {
	rest *api.RESTClient
}

// NewReadOnlyClient resolves credentials the same way gh does
// (GH_TOKEN/GITHUB_TOKEN, then gh's own stored auth).
func NewReadOnlyClient() (*ReadOnlyClient, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	return &ReadOnlyClient{rest: rest}, nil
}

// Get fetches a single resource and decodes it into resp.
func (c *ReadOnlyClient) Get(path string, resp any) error {
	return c.rest.Get(path, resp)
}

// GetRaw fetches path and returns the raw response body, for endpoints
// that don't return JSON (e.g. Actions job logs).
func (c *ReadOnlyClient) GetRaw(path string) ([]byte, error) {
	resp, err := c.rest.Request(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

// GetPaginated follows the Link: rel="next" header starting at path,
// decoding each page as []T and appending to the result, until either
// there is no next page or limit items have been collected (limit<=0
// means no limit).
func GetPaginated[T any](c *ReadOnlyClient, path string, limit int) ([]T, error) {
	var all []T
	requestPath := path
	for {
		resp, err := c.rest.Request(http.MethodGet, requestPath, nil)
		if err != nil {
			return nil, err
		}

		var page []T
		err = json.NewDecoder(resp.Body).Decode(&page)
		closeErr := resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}

		all = append(all, page...)
		if limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}

		next, ok := nextPage(resp)
		if !ok {
			return all, nil
		}
		requestPath = next
	}
}

func nextPage(resp *http.Response) (string, bool) {
	for _, m := range linkRE.FindAllStringSubmatch(resp.Header.Get("Link"), -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false
}
