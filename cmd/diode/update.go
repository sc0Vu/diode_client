package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kierdavis/ansi"

	"github.com/diodechain/go-update"
	"github.com/diodechain/go-update/progress"
	"github.com/diodechain/go-update/stores/github"
)

func doUpdate() {
	m := &update.Manager{
		Command: "diode",
		Store: &github.Store{
			Owner:   "diodechain",
			Repo:    "diode_go_client",
			Version: version,
		},
	}

	if runtime.GOOS == "windows" {
		m.Command += ".exe"
	}

	tarball, ok := download(m)
	if !ok {
		return
	}

	// searching for binary in path
	bin, err := exec.LookPath(m.Command)
	if err != nil {
		// just update local file
		bin = os.Args[0]
	}

	dir := filepath.Dir(bin)
	if err := m.InstallTo(tarball, dir); err != nil {
		log.Fatalf("error installing: %s", err)
	}

	cmd := path.Join(dir, m.Command)
	fmt.Printf("Updated, restarting %s...\n", cmd)

	err = update.Restart(cmd)
	if err != nil {
		log.Fatalf("error restarting %s: %v", cmd, err)
	}
}

func download(m *update.Manager) (string, bool) {
	ansi.HideCursor()
	defer ansi.ShowCursor()

	log.Printf("Checking for updates...")

	// fetch the new releases
	releases, err := m.LatestReleases()
	if err != nil {
		log.Printf("error fetching releases: %s", err)
		return "", false
	}

	// no updates
	if len(releases) == 0 {
		log.Print("no updates")
		return "", false
	}

	// latest release
	latest := releases[0]
	log.Printf("found version %s > %s\n", latest.Version, version)

	a := latest.FindZip(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		log.Printf("no binary for your system (%s_%s)", runtime.GOOS, runtime.GOARCH)
		return "", false
	}

	// whitespace
	fmt.Println()

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		log.Fatalf("error downloading: %s", err)
	}

	return tarball, true
}
