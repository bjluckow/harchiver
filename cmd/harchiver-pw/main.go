package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bjluckow/harchiver/internal/browser"
)

func main() {
	var (
		harPath  = flag.String("har", "out.har", "Output HAR file")
		timeout  = flag.Duration("timeout", 15*time.Second, "Navigation timeout")
		headless = flag.Bool("headless", true, "Run browser headless")
	)

	flag.Parse()

	var urls []string

	// 1. Positional args
	if args := flag.Args(); len(args) > 0 {
		urls = append(urls, args...)
	} else {
		// 2. stdin
		stat, err := os.Stdin.Stat()
		if err != nil {
			log.Fatalf("stdin stat failed: %v", err)
		}

		if (stat.Mode() & os.ModeCharDevice) != 0 {
			log.Fatal("no URLs provided (args or stdin)")
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			urls = append(urls, line)
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("reading stdin: %v", err)
		}
	}

	if len(urls) == 0 {
		log.Fatal("no URLs provided")
	}

	fmt.Printf("Fetching %d URL(s) → %s\n", len(urls), *harPath)

	err := browser.FetchWithHAR(browser.FetchOptions{
		URLs:     urls,
		HARPath:  *harPath,
		Timeout:  *timeout,
		Headless: *headless,
	})
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	fmt.Println("HAR written successfully")
}
