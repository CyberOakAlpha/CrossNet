package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/CyberOakAlpha/CrossNet/internal/osdetect"
	"github.com/CyberOakAlpha/CrossNet/internal/scanner"
)

const (
	version = "1.0.0"
	banner  = `
  ______                      _   _      _
 / _____)                    | \ | |    | |
| /      ____ ___   ___  ___ |  \| | ___| |_
| |     / ___) _ \ / __|/ __)| . ' |/ _ \ __|
| \____| |  | |_| |__ |__ |  |\  |  __/ |_
 \______)_|   \___/|___/___/|_| \_|\___|\__|

    Network Discovery Tool v%s
    Ping & ARP Scanner for Windows/Linux
    https://github.com/CyberOakAlpha/CrossNet
`
)

type Config struct {
	network     string
	timeout     time.Duration
	threads     int
	scanType    string
	showHelp    bool
	showVersion bool
	verbose     bool
	outputFile  string
}

func main() {
	config := parseFlags()

	if config.showHelp {
		showHelp()
		return
	}

	if config.showVersion {
		fmt.Printf("CrossNet v%s\n", version)
		return
	}

	fmt.Printf(banner, version)
	fmt.Printf("Operating System: %s\n", osdetect.GetOSString())
	fmt.Printf("Scan Type: %s\n", config.scanType)
	fmt.Printf("Network: %s\n", config.network)
	fmt.Printf("Threads: %d\n", config.threads)
	fmt.Printf("Timeout: %v\n\n", config.timeout)

	switch strings.ToLower(config.scanType) {
	case "ping":
		runPingScan(config)
	case "arp":
		runARPScan(config)
	case "both":
		runPingScan(config)
		fmt.Println()
		runARPScan(config)
	default:
		fmt.Printf("Error: Invalid scan type '%s'. Use 'ping', 'arp', or 'both'\n", config.scanType)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.network, "network", "192.168.1.0/24", "Network to scan (CIDR notation)")
	flag.StringVar(&config.network, "n", "192.168.1.0/24", "Network to scan (CIDR notation) - short")
	flag.DurationVar(&config.timeout, "timeout", 2*time.Second, "Timeout for ping requests")
	flag.DurationVar(&config.timeout, "t", 2*time.Second, "Timeout for ping requests - short")
	flag.IntVar(&config.threads, "threads", 50, "Number of concurrent threads")
	flag.IntVar(&config.threads, "T", 50, "Number of concurrent threads - short")
	flag.StringVar(&config.scanType, "scan", "both", "Scan type: ping, arp, or both")
	flag.StringVar(&config.scanType, "s", "both", "Scan type: ping, arp, or both - short")
	flag.BoolVar(&config.showHelp, "help", false, "Show help message")
	flag.BoolVar(&config.showHelp, "h", false, "Show help message - short")
	flag.BoolVar(&config.showVersion, "version", false, "Show version")
	flag.BoolVar(&config.showVersion, "v", false, "Show version - short")
	flag.BoolVar(&config.verbose, "verbose", false, "Verbose output")
	flag.StringVar(&config.outputFile, "output", "", "Output file (optional)")
	flag.StringVar(&config.outputFile, "o", "", "Output file (optional) - short")

	flag.Parse()
	return config
}

func showHelp() {
	fmt.Printf(banner, version)
	fmt.Println("USAGE:")
	fmt.Println("  crossnet [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -n, --network    Network to scan (CIDR notation) [default: 192.168.1.0/24]")
	fmt.Println("  -s, --scan       Scan type: ping, arp, or both [default: both]")
	fmt.Println("  -t, --timeout    Timeout for ping requests [default: 2s]")
	fmt.Println("  -T, --threads    Number of concurrent threads [default: 50]")
	fmt.Println("  -o, --output     Output file (optional)")
	fmt.Println("  -v, --version    Show version")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println("      --verbose    Verbose output")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  crossnet -n 192.168.0.0/24 -s ping")
	fmt.Println("  crossnet -n 10.0.0.0/24 -s arp -T 100")
	fmt.Println("  crossnet -n 172.16.1.0/24 -s both -o results.txt")
	fmt.Println()
}

func runPingScan(config Config) {
	fmt.Println("=== PING SCAN RESULTS ===")

	pingScanner := scanner.NewPingScanner(config.timeout, config.threads)
	results, err := pingScanner.ScanRange(config.network)
	if err != nil {
		fmt.Printf("Error running ping scan: %v\n", err)
		return
	}

	aliveCount := 0
	fmt.Printf("%-15s %-10s %-8s %-30s\n", "IP Address", "Status", "RTT", "Hostname")
	fmt.Println(strings.Repeat("-", 70))

	for _, result := range results {
		if result.Alive {
			aliveCount++
			status := "UP"
			rtt := result.RTT.Truncate(time.Millisecond).String()
			hostname := result.Hostname
			if hostname == "" {
				hostname = "N/A"
			}
			fmt.Printf("%-15s %-10s %-8s %-30s\n", result.IP, status, rtt, hostname)
		} else if config.verbose {
			status := "DOWN"
			fmt.Printf("%-15s %-10s %-8s %-30s\n", result.IP, status, "N/A", "N/A")
		}
	}

	fmt.Printf("\nPing scan completed. %d/%d hosts are alive.\n", aliveCount, len(results))
}

func runARPScan(config Config) {
	fmt.Println("=== ARP SCAN RESULTS ===")

	arpScanner := scanner.NewARPScanner(config.threads)

	fmt.Println("Getting ARP table...")
	arpEntries, err := arpScanner.GetARPTable()
	if err != nil {
		fmt.Printf("Error getting ARP table: %v\n", err)
	} else {
		if len(arpEntries) > 0 {
			fmt.Printf("%-15s %-18s %-10s %-30s\n", "IP Address", "MAC Address", "Status", "Hostname")
			fmt.Println(strings.Repeat("-", 80))

			for _, entry := range arpEntries {
				hostname := entry.Hostname
				if hostname == "" {
					hostname = "N/A"
				}
				status := "CACHED"
				fmt.Printf("%-15s %-18s %-10s %-30s\n", entry.IP, entry.MAC, status, hostname)
			}
			fmt.Printf("\nFound %d entries in ARP table.\n", len(arpEntries))
		} else {
			fmt.Println("No entries found in ARP table.")
		}
	}

	fmt.Println("\nScanning network for active devices...")
	networkEntries, err := arpScanner.ScanNetwork(config.network)
	if err != nil {
		fmt.Printf("Error running network ARP scan: %v\n", err)
		return
	}

	if len(networkEntries) > 0 {
		fmt.Printf("\n%-15s %-18s %-10s %-30s\n", "IP Address", "MAC Address", "Status", "Hostname")
		fmt.Println(strings.Repeat("-", 80))

		for _, entry := range networkEntries {
			hostname := entry.Hostname
			if hostname == "" {
				hostname = "N/A"
			}
			status := "ACTIVE"
			fmt.Printf("%-15s %-18s %-10s %-30s\n", entry.IP, entry.MAC, status, hostname)
		}
		fmt.Printf("\nNetwork scan completed. Found %d active devices.\n", len(networkEntries))
	} else {
		fmt.Println("No active devices found in network scan.")
	}
}