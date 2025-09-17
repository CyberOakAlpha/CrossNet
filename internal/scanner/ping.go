package scanner

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CyberOakAlpha/CrossNet/internal/osdetect"
)

type PingResult struct {
	IP       string
	Hostname string
	Alive    bool
	RTT      time.Duration
	Error    string
}

type PingScanner struct {
	timeout  time.Duration
	threads  int
	protocol string
}

func NewPingScanner(timeout time.Duration, threads int) *PingScanner {
	return &PingScanner{
		timeout: timeout,
		threads: threads,
	}
}

func (ps *PingScanner) ScanRange(network string) ([]PingResult, error) {
	ips, err := generateIPRange(network)
	if err != nil {
		return nil, err
	}

	results := make([]PingResult, 0, len(ips))
	resultChan := make(chan PingResult, len(ips))
	semaphore := make(chan struct{}, ps.threads)

	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := ps.pingHost(ipAddr)
			resultChan <- result
		}(ip)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

func (ps *PingScanner) pingHost(ip string) PingResult {
	result := PingResult{
		IP:    ip,
		Alive: false,
	}

	hostname, _ := net.LookupAddr(ip)
	if len(hostname) > 0 {
		result.Hostname = hostname[0]
	}

	var cmd *exec.Cmd
	start := time.Now()

	switch osdetect.DetectOS() {
	case osdetect.Windows:
		cmd = exec.Command("ping", "-n", "1", "-w", strconv.Itoa(int(ps.timeout.Milliseconds())), ip)
	case osdetect.Linux, osdetect.Darwin:
		timeoutSec := int(ps.timeout.Seconds())
		if timeoutSec == 0 {
			timeoutSec = 1
		}
		cmd = exec.Command("ping", "-c", "1", "-W", strconv.Itoa(timeoutSec), ip)
	default:
		result.Error = "Unsupported operating system"
		return result
	}

	output, err := cmd.Output()
	result.RTT = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		return result
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "TTL=") || strings.Contains(outputStr, "ttl=") ||
		strings.Contains(outputStr, "time=") || strings.Contains(outputStr, "time<") {
		result.Alive = true
	}

	return result
}

func generateIPRange(network string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %v", err)
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}