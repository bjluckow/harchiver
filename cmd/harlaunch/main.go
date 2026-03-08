package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
)

func main() {
	var (
		execPath = flag.String("exe", "", "Path to Chrome/Chromium executable")
		port     = flag.Int("port", 9222, "Remote debugging port")
		headless = flag.Bool("headless", false, "Run headless")
	)
	flag.Parse()

	if *execPath == "" {
		log.Fatal("-exe is required")
	}

	// Check if port is already in use
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("port %d already in use", *port)
	}
	ln.Close()

	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", *port),
		"--no-first-run",
		"--no-default-browser-check",
	}

	if *headless {
		args = append(args, "--headless=new")
	}

	// Append any extra args after --
	args = append(args, flag.Args()...)

	cmd := exec.Command(*execPath, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("start chrome: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Chrome started (PID %d)\n", cmd.Process.Pid)
	fmt.Fprintf(os.Stderr, "CDP endpoint: ws://127.0.0.1:%d/\n", *port)

	// Also print just the endpoint to stdout for piping
	fmt.Printf("ws://127.0.0.1:%d/\n", *port)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	fmt.Fprintln(os.Stderr, "Shutting down Chrome...")
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}
