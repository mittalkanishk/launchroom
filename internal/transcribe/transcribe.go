// Package transcribe turns a voice note into text with OpenAI's transcription API, so the agents can
// act on what people say in the room. Chat apps hand the agents a voice note as an audio file
// (Telegram's is voice.ogg), which Claude Code can't listen to.
package transcribe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// maxBytes is the largest file the API accepts.
const maxBytes = 25 << 20

// Config is the API account and model.
type Config struct {
	APIKey  string
	Model   string // for example gpt-4o-mini-transcribe
	BaseURL string // https://api.openai.com/v1; tests point it at a local server
}

// Client sends audio files to the transcription API.
type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config, httpClient *http.Client) *Client {
	return &Client{
		cfg:  cfg,
		http: httpClient,
	}
}

// Transcribe returns what is said in the audio file at path.
func (c *Client) Transcribe(ctx context.Context, path string) (string, error) {
	if c.cfg.APIKey == "" {
		return "", errors.New("OPENAI_API_KEY is not set")
	}
	body, contentType, err := c.form(path)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/audio/transcriptions", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("transcribe: %w", err)
	}
	defer resp.Body.Close()

	var reply struct {
		Text  string `json:"text"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return "", fmt.Errorf("transcribe: %s, with a reply that isn't JSON", resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transcribe: %s %s", resp.Status, reply.Error.Message)
	}
	return strings.TrimSpace(reply.Text), nil
}

// audioFormats are the file signatures of the formats voice notes arrive in, and the extension the
// API needs to read each. The API tells a file's format from its name, but Switch saves attachments
// without an extension (`…-attachment`), so the format is read from the content instead.
var audioFormats = []struct {
	offset    int
	signature string
	extension string
}{
	{0, "OggS", ".ogg"},
	{0, "RIFF", ".wav"},
	{0, "fLaC", ".flac"},
	{0, "ID3", ".mp3"},
	{0, "\x1a\x45\xdf\xa3", ".webm"},
	{4, "ftyp", ".m4a"},
}

// uploadName is the file's name with the extension its content calls for. A name that already ends
// in a known extension, or content it doesn't recognise, is left as it is.
func uploadName(path string, content []byte) string {
	name := filepath.Base(path)
	for _, f := range audioFormats {
		if strings.EqualFold(filepath.Ext(name), f.extension) {
			return name
		}
	}
	for _, f := range audioFormats {
		end := f.offset + len(f.signature)
		if len(content) >= end && string(content[f.offset:end]) == f.signature {
			return strings.TrimSuffix(name, filepath.Ext(name)) + f.extension
		}
	}
	return name
}

// form is the multipart request body: the model, and the audio file under a name whose extension
// tells the API its format.
func (c *Client) form(path string) (io.Reader, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	if info.Size() > maxBytes {
		return nil, "", fmt.Errorf("%s is %d MB; the limit is 25 MB", path, info.Size()>>20)
	}
	audio, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer audio.Close()

	content, err := io.ReadAll(audio)
	if err != nil {
		return nil, "", err
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("model", c.cfg.Model); err != nil {
		return nil, "", err
	}
	part, err := w.CreateFormFile("file", uploadName(path, content))
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(content); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &body, w.FormDataContentType(), nil
}
