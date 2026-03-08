package browser

import (
	"fmt"
	"net"
	"os"
	"os/exec"
)

type Instance struct {
	Cmd  *exec.Cmd
	Port int
}

func Launch(options *LaunchOptions) (*Instance, error) {
	port := options.Port
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("port %d in use: %w", port, err)
	}
	ln.Close()

	cmd := exec.Command(options.ExecPath, options.Args()...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start chrome: %w", err)
	}

	return &Instance{Cmd: cmd, Port: port}, nil
}

func (i *Instance) Endpoint() string {
	return fmt.Sprintf("ws://127.0.0.1:%d/", i.Port)
}

func (i *Instance) Stop() error {
	if i.Cmd.Process != nil {
		i.Cmd.Process.Signal(os.Interrupt)
		return i.Cmd.Wait()
	}
	return nil
}
