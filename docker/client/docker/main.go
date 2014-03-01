package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dotcloud/docker/api"
	"github.com/dotcloud/docker/docker/client"
	flag "github.com/dotcloud/docker/pkg/mflag"
	"github.com/dotcloud/docker/pkg/opts"
)

func main() {
	var (
		flDebug = flag.Bool([]string{"D", "-debug"}, false, "Enable debug mode")
		flHosts = opts.NewListOpts(api.ValidateHost)
	)
	flag.Var(&flHosts, []string{"H", "-host"}, "tcp://host:port, unix://path/to/socket, fd://* or fd://socketfd to use in daemon mode. Multiple sockets can be specified")

	flag.Parse()

	if flHosts.Len() == 0 {
		defaultHost := os.Getenv("DOCKER_HOST")

		if defaultHost == "" {
			// If we do not have a host, default to unix socket
			defaultHost = fmt.Sprintf("unix://%s", api.DEFAULTUNIXSOCKET)
		}
		if _, err := api.ValidateHost(defaultHost); err != nil {
			log.Fatal(err)
		}
		flHosts.Set(defaultHost)
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}
	client.Run(flHosts)
}
