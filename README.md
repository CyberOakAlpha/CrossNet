# CrossNet

A fast and efficient network discovery tool written in Go that performs ping and ARP scans on Windows and Linux systems.

## Features

- **Cross-platform**: Works on Windows, Linux, and macOS
- **Ping scanning**: Fast ICMP ping sweep with customizable timeout
- **ARP scanning**: ARP table enumeration and network scanning
- **Multi-threaded**: Concurrent scanning for better performance
- **Hostname resolution**: Automatic reverse DNS lookup
- **Flexible output**: Console output with optional file export

## Installation

### Build from source

```bash
git clone https://github.com/CyberOakAlpha/CrossNet.git
cd CrossNet
go build -o crossnet cmd/crossnet/main.go
```

### Download binary

Download the latest release from the [releases page](https://github.com/CyberOakAlpha/CrossNet/releases).

## Usage

### Basic usage

```bash
# Scan default network (192.168.1.0/24) with both ping and ARP
./crossnet

# Ping scan only
./crossnet -s ping -n 192.168.0.0/24

# ARP scan only
./crossnet -s arp -n 10.0.0.0/24

# Custom settings
./crossnet -n 172.16.1.0/24 -s both -T 100 -t 5s
```

### Command line options

```
-n, --network    Network to scan (CIDR notation) [default: 192.168.1.0/24]
-s, --scan       Scan type: ping, arp, or both [default: both]
-t, --timeout    Timeout for ping requests [default: 2s]
-T, --threads    Number of concurrent threads [default: 50]
-o, --output     Output file (optional)
-v, --version    Show version
-h, --help       Show help message
    --verbose    Verbose output (shows offline hosts)
```

### Examples

```bash
# Quick ping scan of local network
./crossnet -s ping

# Comprehensive scan with high thread count
./crossnet -n 192.168.0.0/24 -s both -T 200

# Save results to file
./crossnet -n 10.0.0.0/16 -o network_scan.txt

# Verbose output showing all hosts
./crossnet --verbose
```

## Sample Output

```
  ______                      _   _      _
 / _____)                    | \ | |    | |
| /      ____ ___   ___  ___ |  \| | ___| |_
| |     / ___) _ \ / __|/ __)| . ' |/ _ \ __|
| \____| |  | |_| |__ |__ |  |\  |  __/ |_
 \______)_|   \___/|___/___/|_| \_|\___|\__|

    Network Discovery Tool v1.0.0
    Ping & ARP Scanner for Windows/Linux

Operating System: linux
Scan Type: both
Network: 192.168.1.0/24
Threads: 50
Timeout: 2s

=== PING SCAN RESULTS ===
IP Address      Status     RTT      Hostname
----------------------------------------------------------------------
192.168.1.1     UP         1ms      router.local
192.168.1.100   UP         5ms      desktop-pc.local
192.168.1.150   UP         2ms      laptop.local

Ping scan completed. 3/254 hosts are alive.

=== ARP SCAN RESULTS ===
Getting ARP table...
IP Address      MAC Address        Status     Hostname
--------------------------------------------------------------------------------
192.168.1.1     AA:BB:CC:DD:EE:FF  CACHED     router.local
192.168.1.100   11:22:33:44:55:66  CACHED     desktop-pc.local
```

## Requirements

- Go 1.19+ (for building from source)
- Administrator/root privileges may be required for some ARP operations
- Network connectivity to target networks

## Supported Operating Systems

- ✅ Windows 10/11
- ✅ Linux (Ubuntu, Debian, CentOS, RHEL)
- ✅ macOS

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Troubleshooting

### Permission issues
- On Linux/macOS, run with `sudo` if you encounter permission errors
- Ensure your user has network access privileges

### Firewall blocking
- Some firewalls may block ICMP ping requests
- ARP scanning may be restricted on some networks
- Try running with different scan types if one method fails

### Performance tuning
- Adjust thread count (`-T`) based on your system capabilities
- Increase timeout (`-t`) for slower networks
- Use smaller network ranges for faster scans

## Similar Tools

CrossNet is inspired by tools like:
- fping
- nmap
- arp-scan
- Advanced IP Scanner