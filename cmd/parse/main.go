package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bjluckow/harchiver/pkg/har"
)

func main() {
	var (
		outPath      = flag.String("out", "", "Output file (required)")
		outPathShort = flag.String("o", "", "Output file (shorthand)")
		extractURLs  = flag.Bool("urls", false, "Extract request URLs")
	)

	flag.Parse()

	// Resolve output path
	output := *outPath
	if output == "" {
		output = *outPathShort
	}
	if output == "" {
		log.Fatal("missing required flag: --out / -o")
	}

	// Collect HAR file paths
	harFiles := collectInputs()
	if len(harFiles) == 0 {
		log.Fatal("no HAR files provided")
	}

	if !*extractURLs {
		log.Fatal("default behavior not implemented (TODO): use --urls")
	}

	if err := extractURLsFromHARs(harFiles, output); err != nil {
		log.Fatalf("parse failed: %v", err)
	}
}

func collectInputs() []string {
	if args := flag.Args(); len(args) > 0 {
		return args
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("stdin stat failed: %v", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}

	var inputs []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		inputs = append(inputs, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("reading stdin: %v", err)
	}

	return inputs
}

func extractURLsFromHARs(paths []string, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, path := range paths {
		if err := extractURLsFromHAR(path, w); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
	}

	return nil
}

func extractURLsFromHAR(path string, w *bufio.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var har har.HttpArchive
	if err := json.NewDecoder(f).Decode(&har); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	for _, e := range har.Log.Entries {
		if e.Request.URL != "" {
			if _, err := w.WriteString(e.Request.URL + "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}
