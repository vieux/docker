package client

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"code.google.com/p/go.net/websocket"

	"github.com/docker/docker/api"
	"github.com/docker/docker/pkg/log"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
)

func (cli *DockerCli) ws(path string, setRawTerminal bool, in io.ReadCloser, stdout, stderr io.Writer, started chan io.Closer) error {
	defer func() {
		if started != nil {
			close(started)
		}
	}()

	ws, err := websocket.Dial(fmt.Sprintf("ws://%s/v%s%s", cli.addr, api.APIVERSION, path), "", "http://localhost/")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("Cannot connect to the Docker daemon. Is 'docker -d' running on this host?")
		}
		// Server hijacks the connection, error 'connection closed' expected
	}

	if started != nil {
		started <- ws
	}

	var receiveStdout chan error

	var oldState *term.State

	if in != nil && setRawTerminal && cli.isTerminalIn && os.Getenv("NORAW") == "" {
		oldState, err = term.SetRawTerminal(cli.inFd)
		if err != nil {
			return err
		}
		defer term.RestoreTerminal(cli.inFd, oldState)
	}

	if stdout != nil || stderr != nil {
		receiveStdout = promise.Go(func() (err error) {
			defer func() {
				if in != nil {
					if setRawTerminal && cli.isTerminalIn {
						term.RestoreTerminal(cli.inFd, oldState)
					}
					// For some reason this Close call blocks on darwin..
					// As the client exists right after, simply discard the close
					// until we find a better solution.
					if runtime.GOOS != "darwin" {
						in.Close()
					}
				}
			}()

			// When TTY is ON, use regular copy
			if setRawTerminal && stdout != nil {
				_, err = io.Copy(stdout, ws)
			} else {
				_, err = stdcopy.StdCopy(stdout, stderr, ws)
			}
			log.Debugf("[hijack] End of stdout")
			return err
		})
	}

	sendStdin := promise.Go(func() error {
		if in != nil {
			io.Copy(ws, in)
			log.Debugf("[ws] End of stdin")
		}
		if err := ws.Close(); err != nil {
			log.Debugf("Couldn't send EOF: %s", err)
		}
		// Discard errors due to pipe interruption
		return nil
	})

	if stdout != nil || stderr != nil {
		if err := <-receiveStdout; err != nil {
			log.Debugf("Error receiveStdout: %s", err)
			return err
		}
	}

	if !cli.isTerminalIn {
		if err := <-sendStdin; err != nil {
			log.Debugf("Error sendStdin: %s", err)
			return err
		}
	}
	return nil
}
