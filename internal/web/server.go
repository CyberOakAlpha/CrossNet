package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/CyberOakAlpha/CrossNet/internal/network"
	"github.com/CyberOakAlpha/CrossNet/internal/scanner"
)

type Server struct {
	port     int
	clients  map[chan ScanEvent]bool
	mutex    sync.RWMutex
	scanning bool
}

type ScanRequest struct {
	Network  string `json:"network"`
	ScanType string `json:"scan_type"`
	Threads  int    `json:"threads"`
	Timeout  int    `json:"timeout"`
}

type ScanEvent struct {
	Type     string      `json:"type"`
	Progress int         `json:"progress,omitempty"`
	Message  string      `json:"message,omitempty"`
	Result   interface{} `json:"result,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type CurrentIPResponse struct {
	Success bool   `json:"success"`
	IP      string `json:"ip,omitempty"`
	Network string `json:"network,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewServer(port int) *Server {
	return &Server{
		port:    port,
		clients: make(map[chan ScanEvent]bool),
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/current-ip", s.handleCurrentIP)
	http.HandleFunc("/api/scan", s.handleScan)
	http.HandleFunc("/api/scan-progress", s.handleScanProgress)
	http.HandleFunc("/api/stop-scan", s.handleStopScan)

	// Serve static files with proper MIME types
	http.HandleFunc("/style.css", s.handleCSS)
	http.HandleFunc("/script.js", s.handleJS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Printf("Starting CrossNet web server on port %d", s.port)
	log.Printf("Open http://localhost:%d in your browser", s.port)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("web", "static", "index.html"))
}

func (s *Server) handleCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	http.ServeFile(w, r, filepath.Join("web", "static", "style.css"))
}

func (s *Server) handleJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeFile(w, r, filepath.Join("web", "static", "script.js"))
}

func (s *Server) handleCurrentIP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ip, network, err := network.GetCurrentIP()
	if err != nil {
		response := CurrentIPResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CurrentIPResponse{
		Success: true,
		IP:      ip,
		Network: network,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.Lock()
	if s.scanning {
		s.mutex.Unlock()
		http.Error(w, "Scan already in progress", http.StatusConflict)
		return
	}
	s.scanning = true
	s.mutex.Unlock()

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.mutex.Lock()
		s.scanning = false
		s.mutex.Unlock()
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})

	go s.runScan(req)
}

func (s *Server) handleScanProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan ScanEvent, 100)

	s.mutex.Lock()
	s.clients[clientChan] = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.clients, clientChan)
		s.mutex.Unlock()
		close(clientChan)
	}()

	for {
		select {
		case event := <-clientChan:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()

			if event.Type == "complete" || event.Type == "error" {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleStopScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.Lock()
	s.scanning = false
	s.mutex.Unlock()

	s.broadcastEvent(ScanEvent{
		Type:    "error",
		Message: "Scan stopped by user",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (s *Server) runScan(req ScanRequest) {
	defer func() {
		s.mutex.Lock()
		s.scanning = false
		s.mutex.Unlock()
	}()

	timeout := time.Duration(req.Timeout) * time.Second

	switch req.ScanType {
	case "ping":
		s.runPingScan(req.Network, timeout, req.Threads)
	case "arp":
		s.runARPScan(req.Network, req.Threads)
	case "both":
		s.runPingScan(req.Network, timeout, req.Threads)
		if s.isScanning() {
			s.runARPScan(req.Network, req.Threads)
		}
	default:
		s.broadcastEvent(ScanEvent{
			Type:  "error",
			Error: "Invalid scan type",
		})
		return
	}

	s.broadcastEvent(ScanEvent{
		Type:    "complete",
		Message: "Scan completed",
	})
}

func (s *Server) runPingScan(network string, timeout time.Duration, threads int) {
	log.Printf("Starting ping scan on network: %s, timeout: %v, threads: %d", network, timeout, threads)
	s.broadcastEvent(ScanEvent{
		Type:     "progress",
		Progress: 0,
		Message:  "Starting ping scan...",
	})

	pingScanner := scanner.NewPingScanner(timeout, threads)
	results, err := pingScanner.ScanRange(network)
	if err != nil {
		log.Printf("Ping scan error: %v", err)
		s.broadcastEvent(ScanEvent{
			Type:  "error",
			Error: fmt.Sprintf("Ping scan failed: %v", err),
		})
		return
	}

	log.Printf("Ping scan completed, got %d total results", len(results))

	processed := 0
	total := len(results)

	aliveCount := 0
	for _, result := range results {
		if !s.isScanning() {
			return
		}

		if result.Alive {
			aliveCount++
			log.Printf("Found alive host: %s (hostname: %s, rtt: %v)", result.IP, result.Hostname, result.RTT)
			s.broadcastEvent(ScanEvent{
				Type:   "result",
				Result: result,
			})
		}

		processed++
		progress := (processed * 100) / total
		s.broadcastEvent(ScanEvent{
			Type:     "progress",
			Progress: progress,
			Message:  fmt.Sprintf("Ping scan progress: %d/%d", processed, total),
		})
	}

	log.Printf("Ping scan finished: found %d alive hosts out of %d total", aliveCount, total)
}

func (s *Server) runARPScan(network string, threads int) {
	log.Printf("Starting ARP scan on network: %s, threads: %d", network, threads)
	s.broadcastEvent(ScanEvent{
		Type:     "progress",
		Progress: 0,
		Message:  "Starting ARP scan...",
	})

	arpScanner := scanner.NewARPScanner(threads)

	log.Printf("Getting ARP table...")
	arpEntries, err := arpScanner.GetARPTable()
	if err != nil {
		log.Printf("ARP table error: %v", err)
	} else {
		log.Printf("Found %d entries in ARP table", len(arpEntries))
		for _, entry := range arpEntries {
			if !s.isScanning() {
				return
			}

			log.Printf("Found ARP entry: %s -> %s (hostname: %s)", entry.IP, entry.MAC, entry.Hostname)
			s.broadcastEvent(ScanEvent{
				Type:   "result",
				Result: entry,
			})
		}
	}

	s.broadcastEvent(ScanEvent{
		Type:     "progress",
		Progress: 50,
		Message:  "Scanning network for active devices...",
	})

	networkEntries, err := arpScanner.ScanNetwork(network)
	if err != nil {
		s.broadcastEvent(ScanEvent{
			Type:  "error",
			Error: fmt.Sprintf("ARP network scan failed: %v", err),
		})
		return
	}

	for _, entry := range networkEntries {
		if !s.isScanning() {
			return
		}

		s.broadcastEvent(ScanEvent{
			Type:   "result",
			Result: entry,
		})
	}

	s.broadcastEvent(ScanEvent{
		Type:     "progress",
		Progress: 100,
		Message:  "ARP scan completed",
	})
}

func (s *Server) broadcastEvent(event ScanEvent) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for client := range s.clients {
		select {
		case client <- event:
		default:
		}
	}
}

func (s *Server) isScanning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.scanning
}