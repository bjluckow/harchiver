package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bjluckow/harchiver/internal/browser"
	"github.com/bjluckow/harchiver/internal/cdp"
	harutil "github.com/bjluckow/harchiver/pkg/har-util"
)

func main() {
	var (
		cdpEndpoint = flag.String("cdp", "", "Chrome DevTools Protocol websocket endpoint (connect to running browser)")
		execPath    = flag.String("exe", "", "Path to a local Chrome installation to launch")
		headless    = flag.Bool("headless", true, "Run launched browser in headless mode")
		output      = flag.String("o", "", "Output file (default: stdout)")
		timeout     = flag.Duration("timeout", 30*time.Second, "Navigation timeout")
	)
	flag.Parse()

	urls, err := validateURLs(parseURLs(flag.Args()))
	if err != nil {
		log.Fatalf("url validation: %v", err)
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if len(urls) > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), *timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Start browser (if CDP+EXE both provided, try CDP -> fallback EXE)
	var bc *browser.Context

	if *cdpEndpoint != "" {
		bc, err = browser.Remote(ctx, *cdpEndpoint)
		if err != nil && *execPath != "" {
			log.Printf("CDP connect failed (%v), falling back to local browser", err)
			bc, err = browser.Launch(ctx, &browser.LaunchOptions{
				ExecPath: *execPath,
				Headless: *headless,
			})
		}
	} else if *execPath != "" {
		bc, err = browser.Launch(ctx, &browser.LaunchOptions{
			ExecPath: *execPath,
			Headless: *headless,
		})
	} else {
		log.Fatal("must specify -cdp, -exe, or both")
	}
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

	if len(urls) > 0 {
		if err := session.Navigate(bc.Ctx, urls); err != nil {
			log.Fatalf("navigate: %v", err)
		}
	} else {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
	}

	if err = harutil.Write(session.HAR(), w); err != nil {
		log.Fatalf("write HAR: %v", err)
	}
}

func parseURLs(args []string) []string {
	if len(args) > 0 {
		return args
	}

	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return nil
	}

	var urls []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}
	return urls
}

func validateURLs(urls []string) ([]string, error) {
	for _, r := range urls {
		u, err := url.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %w", r, err)
		}
		if u.Scheme == "" {
			return nil, fmt.Errorf("invalid URL %q: missing scheme (http/https)", r)
		}
	}
	return urls, nil
}
