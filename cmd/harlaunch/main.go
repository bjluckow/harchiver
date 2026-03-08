package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bjluckow/harchiver/internal/browser"
)

func main() {
	var (
		execPath    = flag.String("exe", "", "Path to Chrome/Chromium executable")
		port        = flag.Int("port", 9222, "Chrome remote debugging port")
		headless    = flag.Bool("headless", false, "Run Chrome in headless mode")
		timeout     = flag.Duration("timeout", 0, "Auto-shutdown after duration (0 = never)")
		detach      = flag.String("detach", "", "Start and exit. Output: 'pid', 'ws', or 'both'")
		doPrintArgs = flag.Bool("output-args", false, "Output Chrome launch arguments and exit")
	)
	flag.Parse()

	if *execPath == "" {
		log.Fatal("-exe is required")
	}

	options := &browser.LaunchOptions{
		ExecPath:  *execPath,
		Port:      *port,
		Headless:  *headless,
		ExtraArgs: flag.Args(),
	}

	if *doPrintArgs {
		fmt.Println(options)
		os.Exit(0)
	}

	inst, err := browser.Launch(options)
	if err != nil {
		log.Fatalf("launch: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Chrome started (PID %d)\n", inst.Cmd.Process.Pid)
	fmt.Fprintf(os.Stderr, "CDP endpoint: %s\n", inst.Endpoint())
	fmt.Println(inst.Endpoint())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	if *detach != "" && *timeout > 0 {
		log.Fatal("-detach and -timeout cannot be used together")
	}

	if *detach != "" {
		switch *detach {
		case "pid":
			fmt.Println(inst.Cmd.Process.Pid)
		case "ws":
			fmt.Println(inst.Endpoint())
		case "both":
			fmt.Printf("%d %s\n", inst.Cmd.Process.Pid, inst.Endpoint())
		default:
			log.Fatalf("invalid -detach value: %s (use pid, ws, or both)", *detach)
		}
		os.Exit(0)
	}

	if *timeout > 0 {
		fmt.Fprintf(os.Stderr, "Timeout: %s\n", timeout.String())
		startTime := time.Now()

		timer := time.After(*timeout)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

	loop:
		for {
			select {
			case <-sigCh:
				break loop
			case <-timer:
				fmt.Fprintln(os.Stderr, "Timeout reached")
				break loop
			case <-ticker.C:
				elapsed := time.Since(startTime).Truncate(time.Second)
				remaining := time.Until(startTime.Add(*timeout)).Truncate(time.Second)
				fmt.Fprintf(os.Stderr, "Timeout: %s elapsed | %s remaining \n", elapsed, remaining)
			}
		}

	} else {
		<-sigCh
	}

	fmt.Fprintln(os.Stderr, "Shutting down Chrome...")
	inst.Stop()
}
