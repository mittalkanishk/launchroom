package sites

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// packageManager is how a site's dependencies are installed and its build is run.
type packageManager struct {
	lockfile string // the file that picks it; empty matches any site
	program  string
	install  []string
	build    []string
}

// packageManagers in the order they are tried: the one whose lockfile the site has, else npm.
// yarn and pnpm run through corepack, which ships with Node.
var packageManagers = []packageManager{
	{lockfile: "yarn.lock", program: "corepack", install: []string{"yarn", "install"}, build: []string{"yarn", "run", "build"}},
	{lockfile: "pnpm-lock.yaml", program: "corepack", install: []string{"pnpm", "install"}, build: []string{"pnpm", "run", "build"}},
	{lockfile: "package-lock.json", program: "npm", install: []string{"ci"}, build: []string{"run", "build"}},
	{lockfile: "", program: "npm", install: []string{"install"}, build: []string{"run", "build"}},
}

// corepackEnv lets corepack fetch the package manager a site asks for without stopping to ask.
var corepackEnv = []string{"COREPACK_ENABLE_DOWNLOAD_PROMPT=0"}

func packageManagerFor(dir string) packageManager {
	for _, pm := range packageManagers {
		if pm.lockfile == "" || exists(filepath.Join(dir, pm.lockfile)) {
			return pm
		}
	}
	return packageManagers[len(packageManagers)-1]
}

// build installs the site's dependencies and builds it, adding the editable deck if the site is one.
// It returns the folder the build wrote, which the site's launch-room.json names.
func (w *Workspace) build(ctx context.Context, dir string) (string, error) {
	m, err := readManifest(dir)
	if err != nil {
		return "", err
	}
	pm := packageManagerFor(dir)
	for _, args := range [][]string{pm.install, pm.build} {
		step := run.Cmd{Dir: dir, Name: pm.program, Args: args, Env: corepackEnv}
		if err := w.runner.Run(ctx, step); err != nil {
			return "", err
		}
	}

	built := filepath.Join(dir, m.buildDir())
	if info, err := os.Stat(built); err != nil || !info.IsDir() {
		return "", fmt.Errorf("the build did not produce %s/", m.buildDir())
	}
	if err := w.exportDeck(ctx, dir, m, built); err != nil {
		return "", err
	}
	return built, nil
}
