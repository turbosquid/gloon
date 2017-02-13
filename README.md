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

Simply copy the executable included in the release to somehwere on your host and you should be ready to go.
