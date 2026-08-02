// Command dsa-save-editor opens a DragonSword Awakening save file and serves a
// local browser UI to inspect and edit it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gfriloux/dragonsword-save-editor/internal/config"
	"github.com/gfriloux/dragonsword-save-editor/internal/domain"
	"github.com/gfriloux/dragonsword-save-editor/internal/web"
)

func main() {
	log.SetFlags(0)
	var (
		addr    = flag.String("addr", "127.0.0.1:0", "listen address (host:port; port 0 picks a free one)")
		noOpen  = flag.Bool("no-open", false, "do not open a browser automatically")
		gameDir = flag.String("game-dir", "", "set and remember the game folder, then start")
		version = flag.Bool("version", false, "print version and exit")
	)
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println("dsa-save-editor", buildVersion)
		return
	}

	if *gameDir != "" {
		if err := (config.Config{GameDir: *gameDir}).Store(); err != nil {
			log.Fatalf("cannot save game folder: %v", err)
		}
		log.Printf("game folder set: %s", *gameDir)
	}

	cat, err := domain.LoadCatalog(domain.DefaultOverridesPath())
	if err != nil {
		log.Fatalf("cannot load item catalog: %v", err)
	}
	srv := web.New(cat)
	defer srv.Close()

	// Power-user convenience: a save path given on the CLI, or auto-detected,
	// is pre-opened so returning users skip the picker. Otherwise the UI drives
	// the first-run game-folder + slot selection.
	if path := savePath(); path != "" {
		if err := srv.Open(path); err != nil {
			log.Printf("could not pre-open %q: %v", path, err)
		} else {
			log.Printf("editing %s", filepath.Base(path))
		}
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	url := fmt.Sprintf("http://%s/", ln.Addr().String())

	httpd := &http.Server{Handler: srv}
	go func() {
		if err := httpd.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}()

	log.Printf("DragonSword Awakening save editor: %s", url)
	if !*noOpen {
		openBrowser(url)
	}
	log.Print("press Ctrl+C to quit")

	waitForSignal()
	log.Println("bye")
}

// savePath returns a save path given on the CLI, or one auto-detected from the
// configured game folder / working directory, or "".
func savePath() string {
	if p := flag.Arg(0); p != "" {
		return p
	}
	return autoDetectSave()
}

func usage() {
	fmt.Fprintf(os.Stderr, `dsa-save-editor — DragonSword Awakening save editor

Usage:
  dsa-save-editor [flags] [path/to/<id>_Slot<N>.db]

On first run the browser UI asks for the game folder and lists your saves; the
chosen folder is remembered. A save path (or -game-dir) may be given to skip
the picker. Close the game before writing changes back.

Flags:
`)
	flag.PrintDefaults()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	// best-effort; ignore errors (headless, etc.)
	_ = cmd.Start()
	time.Sleep(150 * time.Millisecond)
}
