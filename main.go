package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type port int

var errNoPfctlMatch = errors.New("no matches from pfctl")
var identRequestRegexp = regexp.MustCompile(`^(\d+)\s*,\s*(\d+)$`)
var listenPort = ":113"
var pfctlRegexp = regexp.MustCompile(`.*tcp .+:(\d+) \((.+):(\d+)\) -> .+:(\d+).*`)

func main() {
	args := os.Args

	if len(args) < 2 {
		log.Fatal("Must specify a private listening port")
	}

	maybePort := args[1]
	privateListeningPort, err := strconv.Atoi(maybePort)

	if err != nil {
		log.Fatal("not a port!")
	}

	l, err := net.Listen("tcp4", "127.0.0.1"+listenPort)

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		c, err := l.Accept()

		if err != nil {
			c.Close()
			continue
		}

		go handler(c, port(privateListeningPort))
	}
}

// Search the pf state for the connection using the firewall's local port and
// the irc server's remote port, so we can find the related private ip and port.
func searchPfctl(fwLocalPort, ircdRemotePort port) (net.IP, port, error) {
	out, err := exec.Command("/sbin/pfctl", "-ss").Output()

	if err != nil {
		return nil, 0, err
	}

	for _, line := range strings.Split(string(out), "\n") {
		matched := pfctlRegexp.FindStringSubmatch(line)
		if matched != nil &&
			strconv.Itoa(int(fwLocalPort)) == matched[1] &&
			strconv.Itoa(int(ircdRemotePort)) == matched[4] {
			ip := net.ParseIP(matched[2])
			portInt, err := strconv.Atoi(matched[3])

			if ip == nil || err != nil {
				continue
			}

			p := port(portInt)

			return ip, p, nil
		}
	}

	return nil, 0, errNoPfctlMatch
}

func dialPrivateServer(ip net.IP, p port, request string) (string, error) {
	netTimeout := time.Second * 5

	c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip.String(), p), netTimeout)

	if err != nil {
		return "", err
	}
	defer c.Close()

	fmt.Fprint(c, request)

	resp, err := bufio.NewReader(c).ReadString('\n')

	if err != nil {
		return "", err
	}

	return resp, nil
}

func handler(c io.ReadWriteCloser, privateListeningPort port) {
	defer c.Close()
	maybeRequest, err := bufio.NewReader(c).ReadString('\n')

	if err != nil {
		return
	}

	request := strings.TrimSpace(maybeRequest)
	matched := identRequestRegexp.FindStringSubmatch(request)

	if matched == nil {
		respondError(c, request)
		return
	}

	fwLocalPort, _ := strconv.Atoi(matched[1])
	ircdRemotePort, _ := strconv.Atoi(matched[2])

	if fwLocalPort < 1 || fwLocalPort > 65535 || ircdRemotePort < 1 || ircdRemotePort > 65535 {
		respondError(c, request)
		return
	}

	privateIP, privatePort, err := searchPfctl(port(fwLocalPort), port(ircdRemotePort))

	if err != nil {
		respondError(c, request)
		return
	}

	response, err := dialPrivateServer(
		privateIP,
		privateListeningPort,
		fmt.Sprintf("%d , %d\n", privatePort, ircdRemotePort),
	)

	if err != nil {
		respondError(c, request)
		return
	}

	fmt.Fprint(c, response)
}

func respondError(c io.Writer, request string) {
	fmt.Println("responding with error")
	fmt.Fprintln(c, request+":ERROR:NO-USER")
}
