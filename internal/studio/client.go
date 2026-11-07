package studio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Client is the video command's side of studio's API.
type Client struct {
	baseURL string // http://studio:8090
	http    *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL: baseURL,
		http:    httpClient,
	}
}

// Submit sends a request and returns the queued job.
func (c *Client) Submit(ctx context.Context, req Request) (Job, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Job{}, err
	}
	var job Job
	return job, c.call(ctx, http.MethodPost, "/v1/jobs", bytes.NewReader(body), &job)
}

func (c *Client) Job(ctx context.Context, id string) (Job, error) {
	var job Job
	return job, c.call(ctx, http.MethodGet, "/v1/jobs/"+url.PathEscape(id), nil, &job)
}

func (c *Client) Jobs(ctx context.Context, video string) ([]Job, error) {
	var jobs []Job
	return jobs, c.call(ctx, http.MethodGet, "/v1/jobs?video="+url.QueryEscape(video), nil, &jobs)
}

// Download saves one of a video's files to out.
func (c *Client) Download(ctx context.Context, video, rel, out string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/files/"+url.PathEscape(video)+"/"+filepath.ToSlash(rel), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("studio: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("studio: %s: %s", rel, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (c *Client) call(ctx context.Context, method, path string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("studio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var reply errorReply
		json.NewDecoder(resp.Body).Decode(&reply)
		if reply.Error != "" {
			return errors.New(reply.Error)
		}
		return fmt.Errorf("studio: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
