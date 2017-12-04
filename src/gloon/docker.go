package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/miekg/dns"
	. "gloon/record_set"
	"log"
	"regexp"
	"strings"
)

type DockerMonitor struct {
	recs            *RecordSet
	settings        *Settings
	cli             *client.Client
	hostname_filter *regexp.Regexp
}

func NewDockerMonitor(recs *RecordSet, settings *Settings) (dm *DockerMonitor, err error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return
	}
	var hostname_filter *regexp.Regexp
	if settings.HostnameFilter != "" {
		hostname_filter, err = regexp.Compile(settings.HostnameFilter)
		if err != nil {
			return
		}
	}
	dm = &DockerMonitor{recs, settings, cli, hostname_filter}
	return
}

func (dm *DockerMonitor) Run() (err error) {
	log.Printf("Starting docker monitor...")
	// See if we need to create any A records since we've just come up
	containers, err := dm.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}
	for _, container := range containers {
		err := dm.addRecord(container.ID)
		if err != nil {
			log.Printf("Unable to add container IP  %s - %s", container.ID[:10], err.Error())
			continue
		}
	}
	// Loop  on docker events forever, adding and removing records as needed
events:
	for {
		// Handle panics here
		defer func() {
			if r := recover(); r != nil {
				handlePanic(r)
			}
		}()
		ev, ev_err := dm.cli.Events(context.Background(), types.EventsOptions{})
		for {
			select {
			case event := <-ev:
				var err error
				log.Printf("Got event: %s %s %s %s", event.Type, event.Action, event.Status, event.Actor.ID[:10])
				if event.Type == "container" && event.Action == "start" && event.Status == "start" {
					err = dm.addRecord(event.Actor.ID)
				} else if event.Type == "container" && event.Action == "die" && event.Status == "die" {
					err = dm.delRecord(event.Actor.ID)
				}
				if err != nil {
					log.Printf("Unable to process %s event: %s", event.Action, err.Error())
				}
			case err := <-ev_err:
				log.Printf("Got error event: %s", err.Error())
				break events
			}
		}
	}
	return
}

func (dm *DockerMonitor) delRecord(ID string) (err error) {
	r := dm.recs
	cli := dm.cli
	container_json, err := cli.ContainerInspect(context.Background(), ID)
	if err != nil {
		log.Printf("Unable to inspect container %s - %s", ID[:10], err.Error())
		return
	}
	hostname := container_json.Config.Hostname
	log.Printf("Removing A record: %s %s %s", ID[:10], container_json.Name, hostname)
	r.Del(dns.TypeA, hostname)
	return
}

func (dm *DockerMonitor) addRecord(ID string) (err error) {
	cli := dm.cli
	recs := dm.recs
	container_json, err := cli.ContainerInspect(context.Background(), ID)
	if err != nil {
		log.Printf("Unable to inspect container %s - %s", ID[:10], err.Error())
		return
	}
	hostname := container_json.Config.Hostname
	// Only publish non-default hostnames
	if strings.Index(ID, hostname) == 0 {
		log.Printf("Ignoring host %s", hostname)
		return
	}
	if dm.hostname_filter != nil && !dm.hostname_filter.MatchString(hostname) {
		log.Printf("NOTE: hostname %s does not match filter %s. Ignoring.", hostname, dm.settings.HostnameFilter)
		return
	}
	ip := getContainerIpV4(container_json, dm.settings.DockerNetwork)
	if dm.settings.AppendDomain != "" {
		hostname = fmt.Sprintf("%s.%s", hostname, dm.settings.AppendDomain)
	}
	log.Printf("Adding A record: %s %s %s %s", ID[:10], container_json.Name, hostname, ip)
	recs.Put(dns.TypeA, hostname, ip)
	return
}

func getContainerIpV4(data types.ContainerJSON, nw string) (ip string) {
	var ep *network.EndpointSettings
	if nw != "" {
		ep = data.NetworkSettings.Networks[nw]
	}
	// else just pick a random one
	for _, v := range data.NetworkSettings.Networks {
		ep = v
		break
	}
	if ep != nil {
		ip = ep.IPAddress
	}
	return ip
}
