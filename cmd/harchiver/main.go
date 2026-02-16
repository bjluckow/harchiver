package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bjluckow/harchiver/internal/browser"
	"github.com/bjluckow/harchiver/internal/capture"
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

	if *cdpEndpoint == "" && *execPath == "" {
		log.Fatal("must specify -cdp, -exe, or both")
	}

	urls, err := parseAndValidateURLs(flag.Args())
	if err != nil {
		log.Fatalf("url validation: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

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

	var w io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalf("create output: %v", err)
		}
		defer f.Close()
		w = f
	}

	err = capture.Run(bc.Ctx, capture.Options{
		URLs:    urls,
		Output:  w,
		Timeout: *timeout,
	})

	if err != nil {
		log.Fatalf("capture: %v", err)
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

func parseAndValidateURLs(raw []string) ([]string, error) {
	parsed := parseURLs(raw)
	if len(parsed) == 0 {
		return nil, errors.New("no URLs provided")
	}

	return validateURLs(parsed)
}
