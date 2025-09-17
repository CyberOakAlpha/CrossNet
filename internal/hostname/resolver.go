package hostname

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/CyberOakAlpha/CrossNet/internal/osdetect"
)

type HostnameResolver struct {
	cache map[string]string
	mutex sync.RWMutex
}

func NewHostnameResolver() *HostnameResolver {
	return &HostnameResolver{
		cache: make(map[string]string),
	}
}

func (hr *HostnameResolver) Resolve(ip string) string {
	// Check cache first
	hr.mutex.RLock()
	if hostname, exists := hr.cache[ip]; exists {
		hr.mutex.RUnlock()
		return hostname
	}
	hr.mutex.RUnlock()

	// Try multiple resolution methods
	hostname := hr.resolveMultiple(ip)

	// Cache the result (even if empty)
	hr.mutex.Lock()
	hr.cache[ip] = hostname
	hr.mutex.Unlock()

	return hostname
}

func (hr *HostnameResolver) resolveMultiple(ip string) string {
	methods := []func(string) string{
		hr.resolveReverseDNS,
		hr.resolveFromHosts,
		hr.resolveNetBIOS,
		hr.resolveMDNS,
	}

	for _, method := range methods {
		if hostname := method(ip); hostname != "" {
			return hostname
		}
	}

	return ""
}

// Method 1: Standard reverse DNS lookup
func (hr *HostnameResolver) resolveReverseDNS(ip string) string {
	names, err := net.LookupAddr(ip)
	if err == nil && len(names) > 0 {
		hostname := names[0]
		// Remove trailing dot if present
		if strings.HasSuffix(hostname, ".") {
			hostname = hostname[:len(hostname)-1]
		}
		return hostname
	}
	return ""
}

// Method 2: Check /etc/hosts file
func (hr *HostnameResolver) resolveFromHosts(ip string) string {
	if osdetect.DetectOS() == osdetect.Windows {
		return hr.resolveFromWindowsHosts(ip)
	}
	return hr.resolveFromUnixHosts(ip)
}

func (hr *HostnameResolver) resolveFromUnixHosts(ip string) string {
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ip {
			return fields[1]
		}
	}
	return ""
}

func (hr *HostnameResolver) resolveFromWindowsHosts(ip string) string {
	// Windows hosts file location
	cmd := exec.Command("type", "C:\\Windows\\System32\\drivers\\etc\\hosts")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ip {
			return fields[1]
		}
	}
	return ""
}

// Method 3: NetBIOS name resolution (for Windows networks)
func (hr *HostnameResolver) resolveNetBIOS(ip string) string {
	switch osdetect.DetectOS() {
	case osdetect.Linux:
		return hr.resolveNetBIOSLinux(ip)
	case osdetect.Windows:
		return hr.resolveNetBIOSWindows(ip)
	default:
		return ""
	}
}

func (hr *HostnameResolver) resolveNetBIOSLinux(ip string) string {
	// Try nmblookup if available
	cmd := exec.Command("timeout", "3", "nmblookup", "-A", ip)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "<00>") && strings.Contains(line, "UNIQUE") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				name := strings.TrimSpace(fields[0])
				if name != "" && !strings.HasPrefix(name, "__") {
					return name
				}
			}
		}
	}
	return ""
}

func (hr *HostnameResolver) resolveNetBIOSWindows(ip string) string {
	// Use nbtstat on Windows
	cmd := exec.Command("nbtstat", "-A", ip)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "<00>") && strings.Contains(line, "UNIQUE") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return strings.TrimSpace(fields[0])
			}
		}
	}
	return ""
}

// Method 4: mDNS/Bonjour resolution (for Apple devices and modern networks)
func (hr *HostnameResolver) resolveMDNS(ip string) string {
	switch osdetect.DetectOS() {
	case osdetect.Linux:
		return hr.resolveMDNSLinux(ip)
	case osdetect.Darwin:
		return hr.resolveMDNSDarwin(ip)
	default:
		return ""
	}
}

func (hr *HostnameResolver) resolveMDNSLinux(ip string) string {
	// Try avahi-resolve if available
	cmd := exec.Command("timeout", "2", "avahi-resolve", "-a", ip)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	line := strings.TrimSpace(string(output))
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		hostname := fields[1]
		// Remove .local suffix
		if strings.HasSuffix(hostname, ".local") {
			hostname = hostname[:len(hostname)-6]
		}
		return hostname
	}
	return ""
}

func (hr *HostnameResolver) resolveMDNSDarwin(ip string) string {
	// Try dns-sd on macOS
	cmd := exec.Command("dns-sd", "-q", ip, "PTR")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ".local.") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasSuffix(field, ".local.") {
					hostname := field[:len(field)-7] // Remove ".local."
					return hostname
				}
			}
		}
	}
	return ""
}

// Clear the cache
func (hr *HostnameResolver) ClearCache() {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()
	hr.cache = make(map[string]string)
}

// Get cache statistics
func (hr *HostnameResolver) GetCacheStats() (int, []string) {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	var resolved []string
	for ip, hostname := range hr.cache {
		if hostname != "" {
			resolved = append(resolved, fmt.Sprintf("%s -> %s", ip, hostname))
		}
	}

	return len(hr.cache), resolved
}