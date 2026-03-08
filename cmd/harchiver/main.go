package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bjluckow/harchiver/internal/browser"
	"github.com/bjluckow/harchiver/pkg/cdputil"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	var (
		cdpEndpoint = flag.String("cdp", "", "CDP endpoint (connect to running browser)")
		execPath    = flag.String("exe", "", "Chrome executable (launch automatically)")
		headless    = flag.Bool("headless", true, "Run launched browser headless")
		port        = flag.Int("port", 9222, "Debug port for launched Chrome")
		output      = flag.String("out", "", "Output file (default: stdout)")
	)
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		log.Fatal("no URLs provided")
	}
	if *cdpEndpoint == "" && *execPath == "" {
		log.Fatal("must specify -cdp or -exe")
	}

	// Determine CDP endpoint
	var endpoint string
	if *cdpEndpoint != "" {
		endpoint = *cdpEndpoint
	} else {
		options := &browser.LaunchOptions{
			ExecPath:  *execPath,
			Port:      *port,
			Headless:  *headless,
			ExtraArgs: flag.Args(),
		}
		inst, err := browser.Launch(options)
		if err != nil {
			log.Fatalf("launch: %v", err)
		}
		defer inst.Stop()
		endpoint = inst.Endpoint()

		// Give Chrome a moment to start accepting connections
		time.Sleep(time.Second)
	}

	// Connect
	ctx, cancel, err := cdputil.Connect(context.Background(), endpoint)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer cancel()

	// Create a new tab
	tabCtx, tabCancel := chromedp.NewContext(ctx)
	defer tabCancel()

	if err := chromedp.Run(tabCtx); err != nil {
		log.Fatalf("create tab: %v", err)
	}

	targetID := chromedp.FromContext(tabCtx).Target.TargetID

	// Record only this target
	rec := cdputil.NewRecorder()
	rec.ListenTarget(tabCtx, string(targetID))

	if err := chromedp.Run(tabCtx, network.Enable()); err != nil {
		log.Fatalf("enable network: %v", err)
	}

	// Visit each URL
	for _, u := range urls {
		fmt.Fprintf(os.Stderr, "Fetching %s\n", u)
		if err := chromedp.Run(tabCtx,
			chromedp.Navigate(u),
			chromedp.WaitReady("body", chromedp.ByQuery),
		); err != nil {
			log.Printf("navigate %s: %v", u, err)
		}
	}

	// Write HAR
	var w *os.File
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalf("create output: %v", err)
		}
		defer f.Close()
		w = f
	} else {
		w = os.Stdout
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rec.HAR()); err != nil {
		log.Fatalf("write HAR: %v", err)
	}
}
