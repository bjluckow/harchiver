package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	cdputil "github.com/bjluckow/harchiver/pkg/cdputil"
)

func main() {
	var (
		cdpEndpoint = flag.String("cdp", "", "Chrome DevTools Protocol websocket endpoint (connect to running browser)")
		output      = flag.String("out", "", "Output file (default: stdout)")
		compat      = flag.Bool("compat", false, "Check if CDP endpoint is reachable and exit")
	)
	flag.Parse()

	if *compat {
		_, cancel, err := cdputil.Connect(context.Background(), *cdpEndpoint)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cancel()
		fmt.Fprintln(os.Stderr, "OK")
		os.Exit(0)
	}

	if *cdpEndpoint == "" {
		log.Fatal("-cdp <endpoint> is required")
	}

	ctx, cancel, err := cdputil.Connect(context.Background(), *cdpEndpoint)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer cancel()

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

	session := cdputil.NewSession(ctx)
	if err := session.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "Listening... press Ctrl+C to stop and write HAR")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err = enc.Encode(session.HAR()); err != nil {
		log.Fatalf("write HAR: %v", err)
	}
}
