package formspree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mittalkanishk/launch-room/internal/names"
)

// Test sends one signup to the site's form, as a visitor's browser would, and checks that Formspree
// accepted it. It counts toward the project's monthly submissions.
func (p *Project) Test(ctx context.Context, site string) error {
	if err := p.ready(); err != nil {
		return err
	}
	if err := names.Check(site); err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{
		"email":   "form-test+" + site + "@example.com",
		"_gotcha": "", // the spam trap a real form sends empty
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint(site), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("test signup: %w", err)
	}
	defer resp.Body.Close()

	var reply struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return fmt.Errorf("test signup: %s, with a reply that isn't JSON", resp.Status)
	}
	if !reply.OK {
		return fmt.Errorf("test signup failed: %s %s", resp.Status, reply.Error)
	}
	return nil
}
