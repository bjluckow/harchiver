package browser

import (
	"fmt"
	"strings"
)

type LaunchOptions struct {
	ExecPath  string
	Port      int
	Headless  bool
	ExtraArgs []string
}

func (o *LaunchOptions) Args() []string {
	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", o.Port),
		"--no-first-run",
		"--no-default-browser-check",
	}

	if o.Headless {
		args = append(args, "--headless=new")
	}

	args = append(args, o.ExtraArgs...)
	return args
}

func (o *LaunchOptions) String() string {
	return strings.Join(o.Args(), " ")
}
