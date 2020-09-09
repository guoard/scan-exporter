package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
	"github.com/tatsushid/go-fastping"
)

type conf struct {
	Targets []target `yaml:"targets"`
}

type target struct {
	Name   string   `yaml:"name"`
	Period string   `yaml:"period"`
	IP     string   `yaml:"ip"`
	TCP    protocol `yaml:"tcp"`
	UDP    protocol `yaml:"udp"`
	// {tcp,udp}PortsToScan holds all the ports that will be scanned
	// those fields are fielded after having parsed the range given in
	// config file.
	tcpPortsToScan []string
	udpPortsToScan []string
	// those arrays will hold open ports
	tcpPortsOpen []string
	udpPortsOpen []string
}

type protocol struct {
	Name     string
	Range    string `yaml:"range"`
	Expected string `yaml:"expected"`
}

func main() {
	c := conf{}
	confPath := getConfPath(os.Args)
	c.getConf(confPath)

	log.Infof("%d targets found in %s", len(c.Targets), confPath)

	// targetList is an array that will contain each instance of up target found in conf file
	targetList := []target{}
	for i := 0; i < len(c.Targets); i++ {
		t := target{}
		if t.getStatus() {
			// if the target is up, we add it to targetList
			targetList = append(targetList, t)
		} else {
			// else, we log that the target is down
			// maybe we can send a mail or a notification to manually inspect this case ?
			log.Warnf("%s (%s) seems to be down", t.Name, t.IP)
		}
	}

	/*
		from now, we have a valid list of apps to scan in targetList.
		next step is to parse ports ranges for each protocol, and fill
		{tcp,udp}PortsToScan in each app instance in targetList
	*/

	for i := 0; i < len(targetList); i++ {
		t := targetList[i]
		t.parsePorts()
		t.scanApp()
	}
}

// getStatus returns true if the application respond to ping requests
func (t *target) getStatus() bool {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", t.IP)
	if err != nil {
		return false
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		log.Infof("%s RTT: %v\n", addr.String(), rtt)
	}
	if err = p.Run(); err != nil {
		// we can end up here if we do not run the program as sudo...
		return false
	}

	return true
}

// getAddress returns hostname:port format
func (t *target) getAddress(port string) string {
	return t.IP + ":" + port
}

// parsePorts read app scanning range et fill {tcp,udp}PortsToScan
// with required ports.
// FOR NOW it doesn't support other parameters than 'all' and 'reserved'
func (t *target) parsePorts() {
	/*
		parse TCP ports
	*/
	cmd := t.TCP.Range
	switch cmd {
	case "all":
		for port := 1; port <= 65535; port++ {
			t.tcpPortsToScan = append(t.tcpPortsToScan, strconv.Itoa(port))
		}
		return
	case "reserved":
		for port := 1; port <= 1024; port++ {
			t.tcpPortsToScan = append(t.tcpPortsToScan, strconv.Itoa(port))
		}
		return
	}
	/*
		parse UDP ports
	*/
	cmd = t.UDP.Range
	switch cmd {
	case "all":
		for port := 1; port <= 65535; port++ {
			t.udpPortsToScan = append(t.udpPortsToScan, strconv.Itoa(port))
		}
		return
	case "reserved":
		for port := 1; port <= 1024; port++ {
			t.udpPortsToScan = append(t.udpPortsToScan, strconv.Itoa(port))
		}
		return
	}
}

// parsePortsRange returns an array containing all the ports that
// will be scanned
// func (t *target) parsePortsRange(protType string, prot protocol) []string {
// 	var ports = []string{}
// 	switch prot.Range {
// 	// append all ports to the scan list
// 	case "all":
// 		for port := 1; port <= 65535; port++ {

// 		}
// 		return ports
// 	// append reserved ports to the scan list
// 	case "reserved":
// 		for port := 1; port <= 1024; port++ {
// 			ports = append(ports, strconv.Itoa(port))
// 		}
// 		return ports
// 	}

// 	if strings.Contains(a.scanRange, "-") {
// 		// get the list's bounds
// 		content := strings.Split(a.scanRange, "-")
// 		first, err := strconv.Atoi(content[0])
// 		last, err := strconv.Atoi(content[len(content)-1])
// 		if err != nil {
// 			log.Errorf("An error occured while getting ports to scan: %s", err)
// 		}

// 		for port := first; port <= last; port++ {
// 			ports = append(ports, strconv.Itoa(port))
// 		}
// 	}
// 	return ports
// }

func (c *conf) getConf(confFile string) *conf {
	yamlConf, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Errorf("Error while reading %s: %v ", confFile, err)
	}

	if err = yaml.Unmarshal(yamlConf, &c); err != nil {
		log.Errorf("Error while unmarshalling yamlConf: %v", err)
	}

	return c
}

func getConfPath(args []string) string {
	if len(args) > 1 {
		return args[1]
	}
	// default config file
	return "config.yaml"
}

func (t *target) scanApp() {
	// this loop must start goroutines
	for _, port := range t.tcpPortsToScan {
		// get address with host:port format
		address := t.getAddress(port)
		conn, err := net.DialTimeout("tcp", address, 60*time.Second)
		if err != nil {
			fmt.Println(err)
			continue
		}
		conn.Close()
		t.tcpPortsOpen = append(t.tcpPortsOpen, port+"/tcp")
	}
}
