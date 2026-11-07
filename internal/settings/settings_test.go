package settings

import (
	"os"
	"path/filepath"
	"testing"
)

type fakeEnv map[string]string

func (e fakeEnv) Getenv(key string) string { return e[key] }

func TestLoadReadsFileAndEnvironmentWins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.env")
	content := "SITES_DOMAIN=file.example\nFORMSPREE_PROJECT_ID=from-file\nSITES_SRC=/ignored/in/file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path, fakeEnv{"SITES_DOMAIN": "env.example"})
	if err != nil {
		t.Fatal(err)
	}
	if got.SitesDomain != "env.example" {
		t.Errorf("SitesDomain = %q, want the environment's value", got.SitesDomain)
	}
	if got.FormspreeProjectID != "from-file" {
		t.Errorf("FormspreeProjectID = %q, want the file's value", got.FormspreeProjectID)
	}
	if got.SitesSource != "/home/node/sites" {
		t.Errorf("SitesSource = %q, want the default (the file may not set it)", got.SitesSource)
	}
}

func TestLoadWithoutFileUsesDefaults(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "missing.env"), fakeEnv{})
	if err != nil {
		t.Fatal(err)
	}
	if got.SiteTemplates != "/home/node/templates" || got.SitesPublic != "/home/node/public" || got.SitesDomain != "" {
		t.Errorf("defaults = %+v", got)
	}
}
