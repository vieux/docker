package main

import (
	"fmt"
	"github.com/dotcloud/docker"
	"github.com/dotcloud/docker/engine"
	flag "github.com/dotcloud/docker/pkg/mflag"
	"github.com/dotcloud/docker/sysinit"
	"github.com/dotcloud/docker/utils"
	"log"
	"os"
	"strings"
)

var (
	GITCOMMIT string
	VERSION   string
)

func main() {
	if selfPath := utils.SelfPath(); selfPath == "/sbin/init" || selfPath == "/.dockerinit" {
		// Running in init mode
		sysinit.SysInit()
		return
	}

	var (
		flVersion            = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
		flDaemon             = flag.Bool([]string{"d", "-daemon"}, false, "Enable daemon mode")
		flDebug              = flag.Bool([]string{"D", "-debug"}, false, "Enable debug mode")
		flAutoRestart        = flag.Bool([]string{"r", "-restart"}, true, "Restart previously running containers")
		bridgeName           = flag.String([]string{"b", "-bridge"}, "", "Attach containers to a pre-existing network bridge; use 'none' to disable container networking")
		bridgeIp             = flag.String([]string{"#bip", "-bip"}, "", "Use this CIDR notation address for the network bridge's IP, not compatible with -b")
		pidfile              = flag.String([]string{"p", "-pidfile"}, "/var/run/docker.pid", "Path to use for daemon PID file")
		flRoot               = flag.String([]string{"g", "-graph"}, "/var/lib/docker", "Path to use as the root of the docker runtime")
		flEnableCors         = flag.Bool([]string{"#api-enable-cors", "-api-enable-cors"}, false, "Enable CORS headers in the remote API")
		flDns                = docker.NewListOpts(docker.ValidateIp4Address)
		flEnableIptables     = flag.Bool([]string{"#iptables", "-iptables"}, true, "Disable docker's addition of iptables rules")
		flDefaultIp          = flag.String([]string{"#ip", "-ip"}, "0.0.0.0", "Default IP address to use when binding container ports")
		flInterContainerComm = flag.Bool([]string{"#icc", "-icc"}, true, "Enable inter-container communication")
		flGraphDriver        = flag.String([]string{"s", "-storage-driver"}, "", "Force the docker runtime to use a specific storage driver")
		flHosts              = docker.NewListOpts(docker.ValidateHost)
		flMtu                = flag.Int([]string{"#mtu", "-mtu"}, docker.DefaultNetworkMtu, "Set the containers network mtu")
	)
	flag.Var(&flDns, []string{"#dns", "-dns"}, "Force docker to use specific DNS servers")
	flag.Var(&flHosts, []string{"H", "-host"}, "Multiple tcp://host:port or unix://path/to/socket to bind in daemon mode, single connection otherwise")

	flag.Parse()

	if *flVersion {
		showVersion()
		return
	}
	if flHosts.Len() == 0 {
		defaultHost := os.Getenv("DOCKER_HOST")

		if defaultHost == "" || *flDaemon {
			// If we do not have a host, default to unix socket
			defaultHost = fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)
		}
		flHosts.Set(defaultHost)
	}

	if *bridgeName != "" && *bridgeIp != "" {
		log.Fatal("You specified -b & --bip, mutually exclusive options. Please specify only one.")
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}
	docker.GITCOMMIT = GITCOMMIT
	docker.VERSION = VERSION
	if *flDaemon {
		if flag.NArg() != 0 {
			flag.Usage()
			return
		}

		eng, err := engine.New(*flRoot)
		if err != nil {
			log.Fatal(err)
		}
		// Load plugin: httpapi
		job := eng.Job("initapi")
		job.Setenv("Pidfile", *pidfile)
		job.Setenv("Root", *flRoot)
		job.SetenvBool("AutoRestart", *flAutoRestart)
		job.SetenvBool("EnableCors", *flEnableCors)
		job.SetenvList("Dns", flDns.GetAll())
		job.SetenvBool("EnableIptables", *flEnableIptables)
		job.Setenv("BridgeIface", *bridgeName)
		job.Setenv("BridgeIp", *bridgeIp)
		job.Setenv("DefaultIp", *flDefaultIp)
		job.SetenvBool("InterContainerCommunication", *flInterContainerComm)
		job.Setenv("GraphDriver", *flGraphDriver)
		job.SetenvInt("Mtu", *flMtu)
		if err := job.Run(); err != nil {
			log.Fatal(err)
		}
		// Serve api
		job = eng.Job("serveapi", flHosts.GetAll()...)
		job.SetenvBool("Logging", true)
		if err := job.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		if flHosts.Len() > 1 {
			log.Fatal("Please specify only one -H")
		}
		protoAddrParts := strings.SplitN(flHosts.GetAll()[0], "://", 2)
		if err := docker.ParseCommands(protoAddrParts[0], protoAddrParts[1], flag.Args()...); err != nil {
			if sterr, ok := err.(*utils.StatusError); ok {
				if sterr.Status != "" {
					log.Println(sterr.Status)
				}
				os.Exit(sterr.StatusCode)
			}
			log.Fatal(err)
		}
	}
}

func showVersion() {
	fmt.Printf("Docker version %s, build %s\n", VERSION, GITCOMMIT)
}
