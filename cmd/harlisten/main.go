package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/bjluckow/harchiver/internal/browser"
	"github.com/bjluckow/harchiver/internal/cdp"
	harutil "github.com/bjluckow/harchiver/pkg/har-util"
)

func main() {
	var (
		cdpEndpoint = flag.String("cdp", "", "Chrome DevTools Protocol websocket endpoint (connect to running browser)")
		output      = flag.String("o", "", "Output file (default: stdout)")
	)
	flag.Parse()

	if *cdpEndpoint == "" {
		log.Fatal("-cdp <endpoint> is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to remote browser
	bc, err := browser.Remote(ctx, *cdpEndpoint)
	if err != nil {
		log.Fatalf("browser: %v", err)
	}
	defer bc.Cancel()

	// Create output writer
	var w io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalf("create output: %v", err)
		}
		defer f.Close()
		w = f
	}

	session := cdp.NewSession(bc.Ctx)
	if err := session.Start(); err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	if err = harutil.Write(session.HAR(), w); err != nil {
		log.Fatalf("write HAR: %v", err)
	}
}
