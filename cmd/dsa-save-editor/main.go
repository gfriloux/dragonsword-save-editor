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

	"github.com/gfriloux/dragonsword-save-editor/internal/domain"
	"github.com/gfriloux/dragonsword-save-editor/internal/save"
	"github.com/gfriloux/dragonsword-save-editor/internal/web"
)

func main() {
	log.SetFlags(0)
	var (
		addr    = flag.String("addr", "127.0.0.1:0", "listen address (host:port; port 0 picks a free one)")
		noOpen  = flag.Bool("no-open", false, "do not open a browser automatically")
		version = flag.Bool("version", false, "print version and exit")
	)
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println("dsa-save-editor", buildVersion)
		return
	}

	path := flag.Arg(0)
	if path == "" {
		path = autoDetectSave()
		if path == "" {
			usage()
			os.Exit(2)
		}
		log.Printf("auto-detected save: %s", path)
	}

	sv, err := save.Open(path, "")
	if err != nil {
		log.Fatalf("cannot open save %q: %v", path, err)
	}
	defer sv.Close()

	cat, err := domain.LoadCatalog(domain.DefaultOverridesPath())
	if err != nil {
		log.Fatalf("cannot load item catalog: %v", err)
	}
	game := domain.New(sv, cat)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	url := fmt.Sprintf("http://%s/", ln.Addr().String())

	srv := &http.Server{Handler: web.New(game)}
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}()

	log.Printf("DragonSword Awakening save editor: %s", url)
	if !*noOpen {
		openBrowser(url)
	}
	log.Printf("editing %s — press Ctrl+C to quit", filepath.Base(path))

	waitForSignal()
	log.Println("bye")
}

func usage() {
	fmt.Fprintf(os.Stderr, `dsa-save-editor — DragonSword Awakening save editor

Usage:
  dsa-save-editor [flags] [path/to/<id>_Slot<N>.db]

If no path is given, common save locations are searched. Close the game before
writing changes back.

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
