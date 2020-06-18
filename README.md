## go-identproxy

A golang version of a python script I used:

https://github.com/jon2/identproxy

I had to upgrade my pfsense firewall and the python 2.7 package was no longer
available.

I agree with `jon2` that identd is not a useful protocol. But it was fun to do
develop this version and see it work.

In the past I used this [C
program](http://www.clock.org/~fair/opinion/identd.c) to support identd for IRC.

The handy part of a go version is that you can easily target a different os
when compiling, and install it on pfsense.

## Building

To create a binary to install on pfsense:

```
GOOS=freebsd GOARCH=amd64 go build
```

## Setup and Installation

* use `sftp` interactively to upload the binary. Place it in `/usr/local/bin`
* create a NAT rule in pfsense to forward port 113 traffic inbound to the WAN to `127.0.0.1:113`
* create a [shellcmd](https://docs.netgate.com/pfsense/en/latest/development/executing-commands-at-boot-time.html)
  to start `go-identproxy` at startup
  * shellcmd type: `shellcmd`
  * command: `daemon /usr/local/bin/go-identproxy 8113`

Port 8113 is what the [irssi
identd](https://github.com/irssi/scripts.irssi.org/blob/master/scripts/identd.pl)
script is configured to listen on.

See more info on setting up irssi scripts at https://scripts.irssi.org/
