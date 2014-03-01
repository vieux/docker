package client

import (
	"log"
	"os"
	"strings"

	"github.com/dotcloud/docker/api"
	flag "github.com/dotcloud/docker/pkg/mflag"
	"github.com/dotcloud/docker/pkg/opts"
	"github.com/dotcloud/docker/utils"
)

func Run(flHosts opts.ListOpts) {
	if flHosts.Len() > 1 {
		log.Fatal("Please specify only one -H")
	}
	protoAddrParts := strings.SplitN(flHosts.GetAll()[0], "://", 2)
	if err := api.ParseCommands(protoAddrParts[0], protoAddrParts[1], flag.Args()...); err != nil {
		if sterr, ok := err.(*utils.StatusError); ok {
			if sterr.Status != "" {
				log.Println(sterr.Status)
			}
			os.Exit(sterr.StatusCode)
		}
		log.Fatal(err)
	}
}
