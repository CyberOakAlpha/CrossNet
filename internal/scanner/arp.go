package scanner

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/CyberOakAlpha/CrossNet/internal/osdetect"
)

type ARPEntry struct {
	IP          string
	MAC         string
	Hostname    string
	Vendor      string
	Online      bool
	Error       string
}

type ARPScanner struct {
	threads int
}

func NewARPScanner(threads int) *ARPScanner {
	return &ARPScanner{
		threads: threads,
	}
}

func (as *ARPScanner) ScanNetwork(network string) ([]ARPEntry, error) {
	ips, err := generateIPRange(network)
	if err != nil {
		return nil, err
	}

	results := make([]ARPEntry, 0)
	resultChan := make(chan ARPEntry, len(ips))
	semaphore := make(chan struct{}, as.threads)

	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			entry := as.scanHost(ipAddr)
			if entry.MAC != "" {
				resultChan <- entry
			}
		}(ip)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for entry := range resultChan {
		results = append(results, entry)
	}

	return results, nil
}

func (as *ARPScanner) GetARPTable() ([]ARPEntry, error) {
	var cmd *exec.Cmd
	var entries []ARPEntry

	switch osdetect.DetectOS() {
	case osdetect.Windows:
		cmd = exec.Command("arp", "-a")
	case osdetect.Linux, osdetect.Darwin:
		cmd = exec.Command("arp", "-a")
	default:
		return nil, fmt.Errorf("unsupported operating system")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute arp command: %v", err)
	}

	entries = as.parseARPOutput(string(output))
	return entries, nil
}

func (as *ARPScanner) scanHost(ip string) ARPEntry {
	entry := ARPEntry{
		IP:     ip,
		Online: false,
	}

	var cmd *exec.Cmd
	switch osdetect.DetectOS() {
	case osdetect.Windows:
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", ip)
	case osdetect.Linux, osdetect.Darwin:
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	default:
		entry.Error = "Unsupported operating system"
		return entry
	}

	output, err := cmd.Output()
	if err != nil {
		return entry
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "TTL=") || strings.Contains(outputStr, "ttl=") ||
		strings.Contains(outputStr, "time=") || strings.Contains(outputStr, "time<") {
		entry.Online = true

		mac, _ := as.getMACForIP(ip)
		entry.MAC = mac

		hostname, _ := net.LookupAddr(ip)
		if len(hostname) > 0 {
			entry.Hostname = hostname[0]
		}
	}

	return entry
}

func (as *ARPScanner) getMACForIP(ip string) (string, error) {
	var cmd *exec.Cmd

	switch osdetect.DetectOS() {
	case osdetect.Windows:
		cmd = exec.Command("arp", "-a", ip)
	case osdetect.Linux, osdetect.Darwin:
		cmd = exec.Command("arp", "-n", ip)
	default:
		return "", fmt.Errorf("unsupported operating system")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return as.extractMACFromOutput(string(output)), nil
}

func (as *ARPScanner) parseARPOutput(output string) []ARPEntry {
	var entries []ARPEntry
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entry := as.parseARPLine(line)
		if entry.IP != "" && entry.MAC != "" {
			hostname, _ := net.LookupAddr(entry.IP)
			if len(hostname) > 0 {
				entry.Hostname = hostname[0]
			}
			entry.Online = true
			entries = append(entries, entry)
		}
	}

	return entries
}

func (as *ARPScanner) parseARPLine(line string) ARPEntry {
	var entry ARPEntry

	switch osdetect.DetectOS() {
	case osdetect.Windows:
		re := regexp.MustCompile(`\s*(\d+\.\d+\.\d+\.\d+)\s+([0-9a-fA-F-]{17})\s+\w+`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			entry.IP = matches[1]
			entry.MAC = strings.ToUpper(strings.Replace(matches[2], "-", ":", -1))
		}
	case osdetect.Linux, osdetect.Darwin:
		re := regexp.MustCompile(`.*\((\d+\.\d+\.\d+\.\d+)\) at ([0-9a-fA-F:]{17})`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			entry.IP = matches[1]
			entry.MAC = strings.ToUpper(matches[2])
		}
	}

	return entry
}

func (as *ARPScanner) extractMACFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		entry := as.parseARPLine(line)
		if entry.MAC != "" {
			return entry.MAC
		}
	}
	return ""
}