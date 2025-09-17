package network

import (
	"fmt"
	"net"
	"strings"
)

type NetworkInterface struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Network string `json:"network"`
	MAC     string `json:"mac"`
}

func GetCurrentIP() (string, string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", "", fmt.Errorf("failed to get interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if ip.To4() != nil {
				network := calculateNetwork(ip.String())
				return ip.String(), network, nil
			}
		}
	}

	return "", "", fmt.Errorf("no active network interface found")
}

func GetAllNetworkInterfaces() ([]NetworkInterface, error) {
	var interfaces []NetworkInterface

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		netIface := NetworkInterface{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if ip.To4() != nil {
				netIface.IP = ip.String()
				netIface.Network = calculateNetwork(ip.String())
				break
			}
		}

		if netIface.IP != "" {
			interfaces = append(interfaces, netIface)
		}
	}

	return interfaces, nil
}

func calculateNetwork(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ""
	}

	return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
}

func GetDefaultGateway() (string, error) {
	interfaces, err := GetAllNetworkInterfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if !strings.HasPrefix(iface.IP, "127.") && !strings.HasPrefix(iface.IP, "169.254.") {
			parts := strings.Split(iface.IP, ".")
			if len(parts) == 4 {
				gateway := fmt.Sprintf("%s.%s.%s.1", parts[0], parts[1], parts[2])
				return gateway, nil
			}
		}
	}

	return "", fmt.Errorf("unable to determine default gateway")
}

func IsPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ipAddr) {
			return true
		}
	}

	return false
}