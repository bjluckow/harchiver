package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/bjluckow/harchiver/internal/browser"
	"github.com/bjluckow/harchiver/internal/cdp"
)

func main() {
	var (
		cdpEndpoint = flag.String("cdp", "", "Chrome DevTools Protocol websocket endpoint (connect to running browser)")
		isolated    = flag.Bool("isolated", false, "Create isolated browser context")
	)
	flag.Parse()

	urls, err := validateURLs(parseURLs(flag.Args()))
	if err != nil {
		log.Fatalf("url validation: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to remote browser
	bc, err := browser.Remote(ctx, *cdpEndpoint)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer bc.Cancel()

	client := cdp.NewClient(bc.Ctx)
	if *isolated {
		if err := client.NewContext(); err != nil {
			log.Fatalf("new context: %v", err)
		}
	}
	defer client.Close()

	for _, u := range urls {
		if err := client.Navigate(u); err != nil {
			log.Fatalf("navigate %s: %v", u, err)
		}
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
