package renderer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Client is studio's side of the renderer's API.
type Client struct {
	baseURL string // http://renderer:8091
	http    *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL: baseURL,
		http:    httpClient,
	}
}

// Render sends a scene and waits for its video; a scene takes seconds to a few minutes.
func (c *Client) Render(ctx context.Context, source string) (Result, error) {
	body, err := json.Marshal(Request{Source: source})
	if err != nil {
		return Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/render", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("renderer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var reply errorReply
		json.NewDecoder(resp.Body).Decode(&reply)
		if reply.Error != "" {
			return Result{}, errors.New(reply.Error)
		}
		return Result{}, fmt.Errorf("renderer: %s", resp.Status)
	}
	var result Result
	return result, json.NewDecoder(resp.Body).Decode(&result)
}
