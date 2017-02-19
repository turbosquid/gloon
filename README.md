# gloon -  Cross Platform DNS Resolver/Forwarder with Docker Integration, API and More

`gloon` is a forwarding DNS resolver that allows you to dynamically create custom A or A wildcard dns
records via some or all of the following mechanisms:

* In response to docker events. A entries can be created and removed in response to container starts and stops. You can specify
rules that control what containers get A records created, as well as a psuedo-domain under which the containers are placed.
* Via a simple HTTP PUT/DEL methods. Wildcard records can be added and removed (for example, `*.local`).
* In response to changes in a hosts file. `gloon` can monitor any `/etc/hosts` compatible file for changes via polling or
native notification mechanisms and and/remove A records. Wildcards are also supported as above.
* A records can also be added via the command line at startup. Wildcards are supported.

## Installation

The easiest way to get started is via one of the releases. Linux, OSX and Windows are supported. There is only a single executable
with no dependencies. Other platforms and architectures (BSD, RPI, etc) can be easily supported if you build fronm source (see below).

Simply copy the executable included in the release to somehwere on your host and you should be ready to go. In general /usr/local/bin is
is a great place to drop the binary.

## OSX -- Forwarding all traffic to a remote host or vm using a local domain

`gloon` can be run on your OSX host to forward all traffic for a wildcard domain to a specific host or local VM. This saves
you the trouble of updating your /etc/hosts to manage new development hosts. To set this up, run gloon somewhere on your osx
host as follows:

    gloon -l ":5053" -n "*.docker=192.168.1.2" -n "*.*.docker=192.168.1.2"

This will set up gloon to return an ip of 192.168.1.2 for all hosts that match either `*.docker` or `*.*.docker`. Note that you can
change the domain name and ip as needed. You can also add other domains/ips as desired. Do NOT, however. try to use the domain `.local`, as
this wil not resolve properly under osx.

You will also need to set up a file under `/etc/resolvers`  that tells osx to use gloon for any lookups against the `docker` domain. Create a
resolver file as follows in `/etc/resolvers/docker` (or use whatever domain name you chose earlier as the resolver filename):

    nameserver 127.0.0.1
    port 5053

Now try pinging `foo.docker` or `foo.bar.docker`. Ping should show you the ip you set up earlier for the domain. Of course, you'll want something on the
remote host routing traffic to a  matching container on your remote host (see below).

A launchd plist file has been included under the `launchd` directory in this project. Edit this file as needed and copy it into ~/Library/LaunchAgents. This
will start gloon automatically on OSX startup.

## Routing to the the correct container on a VM or remote host

## Building gloon

`gloon` is built with [gb](https://github.com/constabulary/gb), but you can build it with vanilla go as well. All dependencies are vendored under `vendor/`.

To begin, you may find it helpful to set your GOPATH to the project directory and vendor directory. You *must* do this to build with `go build`:

    export GOPATH=$GOPATH:`pwd`:`pwd`/vendor

To build with gb, simply run `gb` in the project root. To build with go, change to `src/gloon` and run `go build`

With either method, set GOOS and GOARCH if desired to cross-compile for specific OS/Arch types.
