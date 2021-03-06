package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

const VERSION = "1.0.3"

func main() {
	log.Printf("I AM GL00N")
	s := Settings{}
	s.Hostnames = []string{}
	app := cli.NewApp()
	app.Name = "gloon"
	app.Usage = "Custom dns resolver with build in docker container support"
	app.UsageText = "gloon [options]"
	app.Version = VERSION
	app.Action = func(c *cli.Context) error {
		if c.StringSlice("hostname") != nil {
			s.Hostnames = c.StringSlice("hostname")
		}
		return appMain(&s)
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "disable-forward",
			Usage:       "Disable request forwarding",
			Destination: &s.DisableForwarding,
		},
		cli.BoolFlag{
			Name:        "disable-docker",
			Usage:       "Disable docker support",
			Destination: &s.DisableDocker,
		},
		cli.BoolFlag{
			Name:        "no-ptr",
			Usage:       "Disable auitomatic PTR record  creation",
			Destination: &s.NoPtr,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "More verbose messaging",
			Destination: &s.Debug,
		},
		cli.StringFlag{
			Name:        "resolvconf, r",
			Value:       "/etc/resolv.conf",
			Usage:       "resolv.conf compatible `FILE` to use for request forwarding",
			Destination: &s.ResolvFile,
		},
		cli.StringFlag{
			Name:        "listen, l",
			Value:       ":53",
			Usage:       "Resolver listens on `ADDR`",
			Destination: &s.ResolverAddr,
		},
		cli.StringFlag{
			Name:        "api-addr, a",
			Value:       "",
			Usage:       "Api http server listens on `ADDR`. Default is no api server",
			Destination: &s.ApiAddr,
		},
		cli.StringFlag{
			Name:        "hostname-filter, f",
			Value:       "",
			Usage:       "Only docker containers with hostnames matching this `REGEX`  will have A records published",
			Destination: &s.HostnameFilter,
		},
		cli.StringFlag{
			Name:        "append-domain, d",
			Value:       "",
			Usage:       "Append `DOMAIN NAME` to all configured A records",
			Destination: &s.AppendDomain,
		},
		cli.StringFlag{
			Name:        "hostfile",
			Value:       "",
			Usage:       "Load up `FILE` at startup and add any records found. Wilcards are supported",
			Destination: &s.Hostfile,
		},
		cli.IntFlag{
			Name:        "reload-interval, i",
			Value:       0,
			Usage:       "Reload hostfile (where applicable) every `SEC` seconds. If unset, default is to try inotify or similiar where available",
			Destination: &s.HostfileReloadInterval,
		},
		cli.StringSliceFlag{
			Name:  "hostname, n",
			Usage: "Add A records in the form of `HOSTNAME=IP`  ",
		},
		cli.StringFlag{
			Name:        "store",
			Value:       "memory",
			Usage:       "Set local dns record storage to `TYPE`. Valid values are 'redis' and 'memory'",
			Destination: &s.Store,
		},
		cli.StringFlag{
			Name:        "store-opts",
			Value:       "",
			Usage:       "Set record store options to `STRING`. Specific values depend on store type. Ex. for redis to set server addr and db: '10.14.2.3:6379,1'",
			Destination: &s.StoreOpts,
		},
		cli.IntFlag{
			Name:        "ttl",
			Value:       3600,
			Usage:       "Returned ttl in `SEC` seconds",
			Destination: &s.Ttl,
		},
		cli.IntFlag{
			Name:        "resolver-timeout",
			Value:       1,
			Usage:       "Pass through resolver timeout in  `SEC` seconds.",
			Destination: &s.ResolverTimeout,
		},
		cli.StringFlag{
			Name:        "docker-network",
			Value:       "",
			Usage:       "Restrict A records to ips found on docker network `NETWORK`",
			Destination: &s.DockerNetwork,
		},
	}

	app.Run(os.Args)
}

func appMain(settings *Settings) (err error) {
	log.Printf("gloon %s starting...", VERSION)
	s, err := NewServer(settings.ResolverAddr, settings)
	if err != nil {
		log.Fatalf("Unable to create server: %s", err.Error())
	}
	if !settings.DisableDocker {
		dm, err := NewDockerMonitor(s.RecordSet, settings)
		if err != nil {
			log.Printf("WARNING: unable to start docker monitor: %s. Docker hostname support will be disabled", err.Error())
		}
		go func() {
			dm.Run()
		}()
	}
	if settings.Hostfile != "" {
		hf := NewHostfile(settings.Hostfile, s.RecordSet, settings.HostfileReloadInterval)
		go func() {
			hf.Run()
		}()
	}

	if settings.ApiAddr != "" {
		go func() {
			RunApiServer(settings, s.RecordSet)
		}()
	}
	err = s.ListenAndServe()
	defer s.Shutdown()
	if err != nil {
		log.Fatalf("Unable to start server: %s", err.Error())
	}
	return
}
