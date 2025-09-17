class CrossNetGUI {
    constructor() {
        this.isScanning = false;
        this.scanResults = [];
        this.eventSource = null;

        this.initializeElements();
        this.bindEvents();
    }

    initializeElements() {
        console.log('Initializing elements...');

        this.elements = {
            getCurrentIPBtn: document.getElementById('get-ip-btn'),
            currentIPInput: document.getElementById('current-ip'),
            networkInput: document.getElementById('network'),
            scanTypeSelect: document.getElementById('scan-type'),
            threadsInput: document.getElementById('threads'),
            timeoutInput: document.getElementById('timeout'),
            scanBtn: document.getElementById('scan-btn'),
            stopBtn: document.getElementById('stop-btn'),
            clearBtn: document.getElementById('clear-btn'),
            status: document.getElementById('status'),
            progressContainer: document.getElementById('progress-container'),
            progress: document.getElementById('progress'),
            progressText: document.getElementById('progress-text'),
            aliveCount: document.getElementById('alive-count'),
            noResults: document.getElementById('no-results'),
            resultsTable: document.getElementById('results-table'),
            resultsBody: document.getElementById('results-body'),
            exportCSV: document.getElementById('export-csv'),
            exportJSON: document.getElementById('export-json')
        };

        // Check if critical elements exist
        if (!this.elements.getCurrentIPBtn) {
            console.error('Get Current IP button not found!');
        } else {
            console.log('Get Current IP button found:', this.elements.getCurrentIPBtn);
        }

        console.log('Elements initialized');
    }

    bindEvents() {
        console.log('Binding events...');

        this.elements.getCurrentIPBtn.addEventListener('click', (e) => {
            console.log('Get Current IP button clicked');
            e.preventDefault();
            this.getCurrentIP();
        });

        this.elements.scanBtn.addEventListener('click', () => this.startScan());
        this.elements.stopBtn.addEventListener('click', () => this.stopScan());
        this.elements.clearBtn.addEventListener('click', () => this.clearResults());
        this.elements.exportCSV.addEventListener('click', () => this.exportResults('csv'));
        this.elements.exportJSON.addEventListener('click', () => this.exportResults('json'));

        this.elements.currentIPInput.addEventListener('input', () => this.updateNetworkFromIP());

        console.log('Events bound successfully');
    }

    async getCurrentIP() {
        try {
            this.elements.getCurrentIPBtn.disabled = true;
            this.elements.getCurrentIPBtn.textContent = 'Detecting...';

            console.log('Fetching current IP...');

            const response = await fetch('/api/current-ip');
            console.log('Response status:', response.status);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            console.log('Response data:', data);

            if (data.success) {
                this.elements.currentIPInput.value = data.ip;
                this.elements.networkInput.value = data.network;
                this.updateStatus('Current IP detected: ' + data.ip, 'complete');
                console.log('IP detected successfully:', data.ip);
            } else {
                this.updateStatus('Failed to detect IP: ' + data.error, 'error');
                console.error('API returned error:', data.error);
            }
        } catch (error) {
            this.updateStatus('Error detecting IP: ' + error.message, 'error');
            console.error('Error in getCurrentIP:', error);
        } finally {
            this.elements.getCurrentIPBtn.disabled = false;
            this.elements.getCurrentIPBtn.textContent = 'Get Current IP';
        }
    }

    updateNetworkFromIP() {
        const ip = this.elements.currentIPInput.value;
        if (ip && this.isValidIP(ip)) {
            const parts = ip.split('.');
            if (parts.length === 4) {
                const network = `${parts[0]}.${parts[1]}.${parts[2]}.0/24`;
                this.elements.networkInput.value = network;
            }
        }
    }

    isValidIP(ip) {
        const regex = /^(\d{1,3}\.){3}\d{1,3}$/;
        if (!regex.test(ip)) return false;

        return ip.split('.').every(part => {
            const num = parseInt(part, 10);
            return num >= 0 && num <= 255;
        });
    }

    async startScan() {
        if (this.isScanning) return;

        const network = this.elements.networkInput.value.trim();
        if (!network) {
            alert('Please enter a network to scan');
            return;
        }

        this.isScanning = true;
        this.updateScanButtons();
        this.clearResults();
        console.log('Starting scan with cleared results');

        const scanConfig = {
            network: network,
            scan_type: this.elements.scanTypeSelect.value,
            threads: parseInt(this.elements.threadsInput.value),
            timeout: parseInt(this.elements.timeoutInput.value)
        };

        try {
            this.updateStatus('Starting scan...', 'scanning');
            this.showProgress();

            const response = await fetch('/api/scan', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(scanConfig)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.startEventStream();
        } catch (error) {
            this.updateStatus('Scan failed: ' + error.message, 'error');
            this.isScanning = false;
            this.updateScanButtons();
            this.hideProgress();
        }
    }

    startEventStream() {
        this.eventSource = new EventSource('/api/scan-progress');

        this.eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleScanUpdate(data);
        };

        this.eventSource.onerror = (error) => {
            console.error('EventSource failed:', error);
            this.eventSource.close();
            this.isScanning = false;
            this.updateScanButtons();
            this.hideProgress();
        };
    }

    handleScanUpdate(data) {
        switch (data.type) {
            case 'progress':
                this.updateProgress(data.progress, data.message);
                break;
            case 'result':
                this.addResult(data.result);
                break;
            case 'complete':
                this.scanComplete(data);
                break;
            case 'error':
                this.updateStatus('Scan error: ' + data.error, 'error');
                this.scanComplete(data);
                break;
        }
    }

    updateProgress(progress, message) {
        this.elements.progress.style.width = progress + '%';
        this.elements.progressText.textContent = progress + '%';
        if (message) {
            this.updateStatus(message, 'scanning');
        }
    }

    addResult(result) {
        console.log('Adding result:', result);
        // Check if this IP already exists to avoid duplicates
        const existingIndex = this.scanResults.findIndex(r => (r.IP || r.ip) === (result.IP || result.ip));
        if (existingIndex >= 0) {
            console.log('Updating existing result for IP:', result.IP || result.ip);
            // Update existing result with new data
            this.scanResults[existingIndex] = result;
        } else {
            console.log('Adding new result for IP:', result.IP || result.ip);
            this.scanResults.push(result);
        }
        console.log('Total results now:', this.scanResults.length);
        this.renderResults();
    }

    scanComplete(data) {
        this.isScanning = false;
        this.updateScanButtons();
        this.hideProgress();

        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }

        const aliveCount = this.scanResults.filter(r => r.Alive || r.Online || r.alive || r.online).length;
        this.updateStatus(`Scan completed. Found ${aliveCount} active devices.`, 'complete');
    }

    stopScan() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }

        fetch('/api/stop-scan', { method: 'POST' })
            .catch(error => console.error('Error stopping scan:', error));

        this.isScanning = false;
        this.updateScanButtons();
        this.hideProgress();
        this.updateStatus('Scan stopped by user', 'error');
    }

    clearResults() {
        this.scanResults = [];
        this.renderResults();
    }

    renderResults() {
        const aliveResults = this.scanResults.filter(r => r.Alive || r.Online || r.alive || r.online);
        this.elements.aliveCount.textContent = aliveResults.length;

        if (this.scanResults.length === 0) {
            this.elements.noResults.style.display = 'block';
            this.elements.resultsTable.style.display = 'none';
            return;
        }

        this.elements.noResults.style.display = 'none';
        this.elements.resultsTable.style.display = 'table';

        this.elements.resultsBody.innerHTML = '';

        aliveResults.forEach(result => {
            const row = document.createElement('tr');

            const isAlive = result.Alive || result.Online || result.alive || result.online;
            const statusClass = isAlive ? 'status-up' : 'status-down';
            const status = (result.Alive || result.alive) ? 'UP' : (result.Online || result.online) ? 'ACTIVE' : 'DOWN';
            const method = (result.RTT !== undefined || result.rtt !== undefined) ? 'PING' : 'ARP';
            const responseTime = (result.RTT || result.rtt) ? this.formatDuration(result.RTT || result.rtt) : 'N/A';

            row.innerHTML = `
                <td>${result.IP || result.ip}</td>
                <td>${result.MAC || result.mac || 'N/A'}</td>
                <td>${result.Hostname || result.hostname || 'N/A'}</td>
                <td class="${statusClass}">${status}</td>
                <td>${responseTime}</td>
                <td>${method}</td>
            `;

            this.elements.resultsBody.appendChild(row);
        });
    }

    formatDuration(nanoseconds) {
        const ms = nanoseconds / 1000000;
        if (ms < 1) {
            return '<1ms';
        }
        return Math.round(ms) + 'ms';
    }

    updateScanButtons() {
        this.elements.scanBtn.disabled = this.isScanning;
        this.elements.stopBtn.disabled = !this.isScanning;
    }

    updateStatus(message, type = 'idle') {
        this.elements.status.textContent = message;
        this.elements.status.className = `status-${type}`;
    }

    showProgress() {
        this.elements.progressContainer.style.display = 'block';
        this.updateProgress(0, '');
    }

    hideProgress() {
        this.elements.progressContainer.style.display = 'none';
    }

    exportResults(format) {
        if (this.scanResults.length === 0) {
            alert('No results to export');
            return;
        }

        const aliveResults = this.scanResults.filter(r => r.Alive || r.Online || r.alive || r.online);

        if (format === 'csv') {
            this.exportCSV(aliveResults);
        } else if (format === 'json') {
            this.exportJSON(aliveResults);
        }
    }

    exportCSV(results) {
        const headers = ['IP Address', 'MAC Address', 'Hostname', 'Status', 'Response Time', 'Method'];
        const csvContent = [
            headers.join(','),
            ...results.map(result => [
                result.IP || result.ip,
                result.MAC || result.mac || '',
                result.Hostname || result.hostname || '',
                (result.Alive || result.alive) ? 'UP' : 'ACTIVE',
                (result.RTT || result.rtt) ? this.formatDuration(result.RTT || result.rtt) : '',
                (result.RTT !== undefined || result.rtt !== undefined) ? 'PING' : 'ARP'
            ].map(field => `"${field}"`).join(','))
        ].join('\n');

        this.downloadFile(csvContent, 'crossnet-results.csv', 'text/csv');
    }

    exportJSON(results) {
        const jsonContent = JSON.stringify(results, null, 2);
        this.downloadFile(jsonContent, 'crossnet-results.json', 'application/json');
    }

    downloadFile(content, filename, contentType) {
        const blob = new Blob([content], { type: contentType });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    }
}

// Initialize the application when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new CrossNetGUI();
});