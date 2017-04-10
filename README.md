# gloon -  Cross Platform DNS Resolver/Forwarder with Docker Integration, API and More

`gloon` is a forwarding DNS resolver that allows you to dynamically create custom A, AAAA,  or A wildcard dns
records via some or all of the following mechanisms:

* In response to docker events. A entries can be created and removed in response to container starts and stops. You can specify
rules that control what containers get A records created, as well as a psuedo-domain under which the containers are placed.
* Via a simple HTTP PUT/DEL methods. Wildcard records can be added and removed (for example, `*.local`).
* In response to changes in a hosts file. `gloon` can monitor any `/etc/hosts` compatible file for changes via polling or
native notification mechanisms and and/remove A records. Wildcards are also supported as above.
* A records can also be added via the command line at startup. Wildcards are supported.
* Support multiple hosts for a single address, or multiple address for a single host (round-robin)
* Creates reverse-dns records auromatically
* optional persistent dns record storage via redis -- multiple gloon instances can share data via redis. Other store types coming


Gloon runs as a single binary with no dependencies by default. Just copy the binary to a host and run it. If you choose to use
redis as a backing store, you'll need to have a redis server available.

## Installation

The easiest way to get started is via one of the releases. Linux, OSX and Windows are supported. There is only a single executable
with no dependencies. Other platforms and architectures (BSD, RPI, etc) can be easily supported if you build fronm source (see below).

Simply copy the executable included in the release to somehwere on your host and you should be ready to go. In general /usr/local/bin is a great place to drop the binary.

## Getting help

Run `gloon -h` to see the available command line options

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

The easiest way to route data to a given container is to run a router container from the Dockerfile in `route/`. This will run gloon, which listens for docker events and creates A records for any containers with a non-default hostname. It also runs nginx, which extracts the hostname from incoming requests and looks up the host container address via gloon. If found, traffic is forwarded to the container. 

To work, the docker socket has to be mapped in to the router container and the router container must in host networking mode. Any container that we want traffic to be routed to has to have a non-default hostname set. There is a `docker-compose.yml` file that can be used to run the router container.

We can combine the request router  with gloon running on OSX as outlined above so that if you point your OSX host browser to http://foo.docker, the request is forwarded to the router on the container host, which looks up the address of the `foo` container, and proxies the request to `foo`.

The nginx configuration will also allow routing traffic to particular ports on a target container via a psuedo-host. For example, say container `foo` is running elasticsearch, which listens for http traffic on port 9200. You can point your browser to http://p9200.foo.docker, and you will be able to reach the listening elasticsearch instance.

This arrangement is convenient in complex environments where exposing ports is impractical due to the number of containers that need to be reachable, such as shared dev/demo environments, qa environments or multi-service development environments.

## Deeper dive: Adding A records

By default, gloon will listen for docker container events, and add and A record, as well as a PTR record for any container with a hostname set. You can set a regex via the `--hostname-filter` flag that can be used to select only matching hostnames to be published. You can disable docker event listening entirely by passing `--disable-docker`.

### Adding records via the http API

Use the `--api-addr` flag to enable the http API server (ex. `--api-addr "127.0.0.1:8080"`). Add or update an A (and ptr) record via PUT:

    curl -XPUT http://localhost:8080/records/A/foo/192.168.1.2 # foo A 192.168.1.2

Remove a record with a corresponding DELETE request:

    curl -XDELETE  http://localhost:8080/records/A/foo
    
You can also add wildcard and double-wildcard records

### Adding records via a hostfile

Use the `--hostfile` flag to have gloon read and monitor a hostfile  to add and remove A records. The hostfile format is the same as `/etc/hosts`, but supports wildcards and double wildcards. gloon will attempt to use native filesystem notifications to check for changes to the hostfile, or you can set a polling interval with `--reload-interval`. Gloon will add new entries where found, and remove entries no longer in the hostfile. An example file `hosts.txt` is included in the project root.

## DNS Forwarding

By default, gloon forwards requests it can't answer to the resolvers configured in /etc/resolv.conf. You can disable forwarding behavior altogether with `--disable-forward`.  You can also specifiy a custom resolv.conf with the `--resolvconf` flag.

It should be generally safe to use gloon in the default resolv.conf file, since gloon tries to be smart enough not to forward unhandled traffic to itself.

## Building gloon

`gloon` is built with [gb](https://github.com/constabulary/gb), but you can build it with vanilla go as well. All dependencies are vendored under `vendor/`.

To begin, you may find it helpful to set your GOPATH to the project directory and vendor directory. You *must* do this to build with `go build`:

    export GOPATH=$GOPATH:`pwd`:`pwd`/vendor

To build with gb, simply run `gb build` in the project root. To build with go, change to `src/gloon` and run `go build`

With either method, set GOOS and GOARCH if desired to cross-compile for specific OS/Arch types.

## Persistent/Shared DNS record storage

By default, gloon stores added dns records in local process memory. However, gloon allows you to use redis as a backing store if desired. When
redis used used, dns records can persist between redis restarts, and multiple gloon processes can use a shared redis server for fault-tolerance.

To enable the redis store pass the store option to gloon:

    gloon --store=redis

By default, gloon will attempt to use a redis server listening on `localhost:6379`. Gloon namespaces any keys it uses in redis with `/gloon` and uses
redis db 0 by default. You can change any of these parameters by passing a comma-delimited set of store-opts when you start gloon. For example:

    gloon --store=redis --store-opts="10.10.0.13:6379,2,test"

would cause gloon to use a redis server at 10.10.0.13 (port 6379), selecting database 2 and using `test` as a namespace. You can start multiple
instances of gloon with the same store opts to share an instance of redis. 

## Known limitations

* We currently only pull the docker container address from the first network found. We should probably add an optional network selector.
* Other record lookups might be desirable (cname, etc)

## Similar projects

* [docker-dns-gen](https://github.com/jderusse/docker-dns-gen)
* [resolvable] (https://github.com/gliderlabs/resolvable)
* [docker-dns] (https://github.com/phensley/docker-dns)
