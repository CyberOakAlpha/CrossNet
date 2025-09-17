# CrossNet

A comprehensive network discovery tool written in Go that performs advanced ping and ARP scans with web GUI and CLI interfaces.

## ✨ Features

### Core Functionality
- **🌐 Cross-platform**: Works on Windows, Linux, and macOS
- **🔍 Dual Interface**: Web GUI and CLI versions
- **⚡ Multi-threaded**: Concurrent scanning for maximum performance
- **📊 Dual Scan Types**: ICMP ping and ARP scanning
- **🏠 Smart Hostname Resolution**: Multiple methods including reverse DNS, NetBIOS, mDNS, and hosts file
- **💾 Export Options**: CSV and JSON export with detailed metrics

### Advanced Features
- **🎯 Intelligent Network Detection**: Auto-detect current IP and network range
- **⚡ Result Caching**: Smart hostname caching for faster repeat scans
- **📈 Real-time Progress**: Live scan progress with WebSocket updates
- **🔧 Configurable**: Adjustable timeout, thread count, and scan parameters
- **📱 Responsive UI**: Modern web interface that works on any device

## 🚀 Quick Start

### Prerequisites

**Go 1.19+ required** for building from source.

**System Dependencies:**
- **Linux**: `ip` command (usually pre-installed)
- **Windows**: No additional dependencies
- **macOS**: No additional dependencies

**Optional Enhancement Tools:**
- **Linux**: `nmblookup` (for NetBIOS resolution), `avahi-resolve` (for mDNS)
- **Windows**: Built-in `nbtstat` and `arp` commands
- **macOS**: Built-in `dns-sd` command

### Installation

#### Option 1: Build from Source
```bash
git clone https://github.com/CyberOakAlpha/CrossNet.git
cd CrossNet
make all  # Builds both CLI and GUI versions
```

#### Option 2: Download Pre-built Binaries
Download from the [releases page](https://github.com/CyberOakAlpha/CrossNet/releases):

**Windows:**
- `crossnet-windows-amd64.exe` + `crossnet-gui-windows-amd64.exe` (64-bit Intel/AMD)
- `crossnet-windows-arm64.exe` + `crossnet-gui-windows-arm64.exe` (ARM64 - Surface Pro X, etc.)

**Linux:**
- `crossnet-linux-amd64` + `crossnet-gui-linux-amd64` (64-bit Intel/AMD)
- `crossnet-linux-arm64` + `crossnet-gui-linux-arm64` (ARM64 - Raspberry Pi 4, etc.)
- `crossnet-linux-armv7` + `crossnet-gui-linux-armv7` (ARMv7 - Raspberry Pi 3, etc.)

**macOS:**
- `crossnet-darwin-amd64` + `crossnet-gui-darwin-amd64` (Intel Mac)
- `crossnet-darwin-arm64` + `crossnet-gui-darwin-arm64` (Apple Silicon - M1/M2/M3)

**Checksums:** `SHA256SUMS` file provided for verification

## 💻 Usage

### Web GUI (Recommended)

```bash
# Start web interface
./crossnet-gui

# Or specify custom port
./crossnet-gui -p 8080
```

Then open http://localhost:8080 in your browser for the full-featured web interface.

### CLI Usage

```bash
# Scan default network with both ping and ARP
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

### Platform-Specific Examples

**Linux/macOS:**
```bash
# Quick ping scan of local network
./crossnet -s ping

# Start web GUI
./crossnet-gui -p 8080

# Comprehensive scan with high thread count
./crossnet -n 192.168.0.0/24 -s both -T 200
```

**Windows:**
```cmd
# Quick ping scan of local network
crossnet-gui.exe

# Start web GUI (open http://localhost:8080)
crossnet-gui.exe -p 8080

# CLI scan
crossnet.exe -n 192.168.1.0/24 -s both
```

**Cross-platform Features:**
```bash
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