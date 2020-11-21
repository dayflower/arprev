package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func usage() {
	fmt.Printf("%s [options] <mac address> <interface> [<network>]\n", os.Args[0])
}

func pingToV4NetworkHosts(network string) error {
	err := exec.Command("nmap", "-sP", network).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "exec nmap command failed.")
		return err
	}

	return nil
}

func pingToV6NetworkHosts(iface string) error {
	err := exec.Command("ping6", "-I", iface, "-c", "3", "ff02::1").Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "exec ping6 command failed.")
		return err
	}

	return nil
}

func findIPV4NetworkByInterface(iface string) (string, error) {
	out, err := exec.Command("ip", "address", "show", "dev", iface, "scope", "global").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "exec ip command failed.")
		return "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		cells := strings.Fields(line)
		if len(cells) < 2 {
			continue
		}

		if cells[0] == "inet" {
			return cells[1], nil
		}
	}

	return "", nil
}

func findIPByArp(mac string, ipv6 bool) (string, error) {
	out, err := exec.Command("ip", "neigh", "show").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "exec ip command failed.")
		return "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		cells := strings.Fields(line)
		if len(cells) < 6 {
			continue
		}

		if cells[4] == mac && cells[5] == "REACHABLE" {
			ip := cells[0]

			if ipv6 {
				if strings.Index(ip, ":") >= 0 {
					return ip, nil
				}
			} else {
				if strings.Index(ip, ".") >= 0 {
					return ip, nil
				}
			}
		}
	}

	return "", nil
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(-1)
	}

	ipv6 := false
	mac := strings.ToLower(os.Args[1])
	iface := os.Args[2]
	network := ""
	if len(os.Args) >= 4 {
		network = os.Args[3]
	}

	ip, err := findIPByArp(mac, ipv6)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// found
	if len(ip) > 0 {
		fmt.Println(ip)
		return
	}

	if network == "" {
		network, err = findIPV4NetworkByInterface(iface)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}

	pingToV4NetworkHosts(network)

	ip, err = findIPByArp(mac, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// found
	if len(ip) > 0 {
		fmt.Println(ip)
		return
	}

	// not found
	os.Exit(1)
}
