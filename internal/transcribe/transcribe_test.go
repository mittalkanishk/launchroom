package transcribe

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeAudio(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "voice.ogg")
	if err := os.WriteFile(path, []byte("OggS fake audio"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTranscribeSendsTheFileAndModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/transcriptions" || r.Header.Get("Authorization") != "Bearer key-123" {
			http.Error(w, `{"error":{"message":"wrong request"}}`, http.StatusBadRequest)
			return
		}
		if r.FormValue("model") != "test-model" {
			http.Error(w, `{"error":{"message":"wrong model"}}`, http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, `{"error":{"message":"no file"}}`, http.StatusBadRequest)
			return
		}
		content, _ := io.ReadAll(file)
		if header.Filename != "voice.ogg" || string(content) != "OggS fake audio" {
			http.Error(w, `{"error":{"message":"wrong file"}}`, http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"text":"  make the headline bigger  "}`))
	}))
	defer server.Close()

	c := NewClient(Config{APIKey: "key-123", Model: "test-model", BaseURL: server.URL}, http.DefaultClient)
	text, err := c.Transcribe(context.Background(), writeAudio(t))
	if err != nil {
		t.Fatal(err)
	}
	if text != "make the headline bigger" {
		t.Errorf("text = %q", text)
	}
}

func TestTranscribeReportsTheAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Incorrect API key provided"}}`))
	}))
	defer server.Close()

	c := NewClient(Config{APIKey: "bad", Model: "m", BaseURL: server.URL}, http.DefaultClient)
	if _, err := c.Transcribe(context.Background(), writeAudio(t)); err == nil || !strings.Contains(err.Error(), "Incorrect API key") {
		t.Errorf("err = %v, want the API's message", err)
	}
}

func TestTranscribeNeedsAKey(t *testing.T) {
	c := NewClient(Config{Model: "m", BaseURL: "http://unused"}, http.DefaultClient)
	if _, err := c.Transcribe(context.Background(), writeAudio(t)); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("err = %v", err)
	}
}

func TestUploadNameGivesSwitchAttachmentsAnExtension(t *testing.T) {
	cases := []struct {
		path    string
		content string
		want    string
	}{
		{"/media/64aba047-attachment", "OggS\x00\x02rest", "64aba047-attachment.ogg"},
		{"/media/abc-attachment", "\x00\x00\x00\x20ftypM4A ", "abc-attachment.m4a"},
		{"/media/abc-attachment", "ID3\x04rest", "abc-attachment.mp3"},
		{"/tmp/voice.ogg", "OggS", "voice.ogg"},
		{"/tmp/note.M4A", "\x00\x00\x00\x20ftyp", "note.M4A"},
		{"/media/unknown-attachment", "hello", "unknown-attachment"},
	}
	for _, c := range cases {
		if got := uploadName(c.path, []byte(c.content)); got != c.want {
			t.Errorf("uploadName(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}
