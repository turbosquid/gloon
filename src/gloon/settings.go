package main

import ()

type Settings struct {
	ResolverAddr           string   // Address resolver listens on
	ApiAddr                string   // Address built-in api http server listens on. Required to enable server
	DisableForwarding      bool     // Disable forwarding of requests to other resolvers when not found. False by default
	DisableDocker          bool     // Diable docker support
	ResolvFile             string   // resolv.conf to use for forwarding. Defaults to /etc/resolv.conf
	HostnameFilter         string   // Only add docker hostnames matching this regex. Defaut is to add all containers w/ a configured hostname
	AppendDomain           string   // Append this domain name to all A records
	Hostfile               string   //Add A records from this file. File supports wildcards
	HostfileReloadInterval int      // Reload hostfile on this interval. If 0 (the default) try using inotify or similiar where vailable
	Hostnames              []string // Hostnames to add from the command line
	Store                  string   // Defaults to memory. "redis" for redis
	StoreOpts              string   // Store-specific options
}
