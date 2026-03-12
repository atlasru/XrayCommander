package xray

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/atlasru/xraycommander/internal/config"
	"github.com/atlasru/xraycommander/pkg/models"
)

// Service manages Xray process
type Service struct {
	configMgr *config.Manager
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	mu        sync.RWMutex
	status    models.ConnectionStatus
	logs      []string
	logMu     sync.Mutex
	onUpdate  func()
}

// NewService creates a new Xray service
func NewService(cm *config.Manager, onUpdate func()) *Service {
	return &Service{
		configMgr: cm,
		status:    models.ConnectionStatus{Connected: false},
		logs:      make([]string, 0, 1000),
		onUpdate:  onUpdate,
	}
}

// GetStatus returns current connection status
func (s *Service) GetStatus() models.ConnectionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// GetLogs returns recent logs
func (s *Service) GetLogs() []string {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	logs := make([]string, len(s.logs))
	copy(logs, s.logs)
	return logs
}

// Start starts Xray with the given profile
func (s *Service) Start(profile *models.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status.Connected {
		return fmt.Errorf("xray is already running")
	}

	xrayPath := s.configMgr.GetXrayPath()
	if xrayPath == "" {
		return fmt.Errorf("xray binary not found")
	}

	// Generate Xray config
	config := profile.ToXrayConfig()
	configPath := s.configMgr.GetRuntimeConfigPath()

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// Start Xray process
	cmd := exec.CommandContext(ctx, xrayPath, "-c", configPath)

	// Set process group for proper termination
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start xray: %w", err)
	}

	s.cmd = cmd
	s.status.Connected = true
	s.status.CurrentProfile = profile
	s.status.StartedAt = time.Now()
	s.status.PID = cmd.Process.Pid

	// Start log readers
	go s.readLogs(stdout)
	go s.readLogs(stderr)

	// Start monitoring
	go s.monitorProcess(cmd, cancel)
	go s.monitorTraffic()

	return nil
}

// Stop stops the Xray process
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.status.Connected {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		// Try graceful termination first
		s.cmd.Process.Signal(syscall.SIGTERM)

		// Wait a bit for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(3 * time.Second):
			// Force kill
			s.cmd.Process.Kill()
		}
	}

	s.status.Connected = false
	s.status.CurrentProfile = nil
	s.status.PID = 0
	s.status.UploadSpeed = 0
	s.status.DownloadSpeed = 0

	return nil
}

// Restart restarts Xray with the current profile
func (s *Service) Restart() error {
	profile := s.status.CurrentProfile
	if err := s.Stop(); err != nil {
		return err
	}

	if profile != nil {
		time.Sleep(500 * time.Millisecond)
		return s.Start(profile)
	}

	return nil
}

func (s *Service) readLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		s.logMu.Lock()
		s.logs = append(s.logs, line)
		if len(s.logs) > 1000 {
			s.logs = s.logs[len(s.logs)-1000:]
		}
		s.logMu.Unlock()

		if s.onUpdate != nil {
			s.onUpdate()
		}
	}
}

func (s *Service) monitorProcess(cmd *exec.Cmd, cancel context.CancelFunc) {
	if err := cmd.Wait(); err != nil {
		s.logMu.Lock()
		s.logs = append(s.logs, fmt.Sprintf("[ERROR] Xray process exited: %v", err))
		s.logMu.Unlock()
	}

	s.mu.Lock()
	s.status.Connected = false
	s.status.CurrentProfile = nil
	s.status.PID = 0
	s.mu.Unlock()

	if s.onUpdate != nil {
		s.onUpdate()
	}
}

func (s *Service) monitorTraffic() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastUpload, lastDownload int64

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			if !s.status.Connected {
				s.mu.RUnlock()
				return
			}
			s.mu.RUnlock()

			// Get traffic stats from API (simplified)
			// In real implementation, use Xray API
			upload, download := s.getTrafficStats()

			s.mu.Lock()
			s.status.UploadSpeed = (upload - lastUpload) / 2
			s.status.DownloadSpeed = (download - lastDownload) / 2
			s.status.TotalUpload = upload
			s.status.TotalDownload = download
			s.mu.Unlock()

			lastUpload = upload
			lastDownload = download

			if s.onUpdate != nil {
				s.onUpdate()
			}
		}
	}
}

func (s *Service) getTrafficStats() (upload, download int64) {
	// Placeholder for actual Xray API integration
	// In production, use Xray's StatsService API
	return 0, 0
}

// TestConnection tests connection to the server
func (s *Service) TestConnection(address string, port int) (int, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), 5*time.Second)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	latency := int(time.Since(start).Milliseconds())
	return latency, nil
}

// IsInstalled checks if Xray is installed
func (s *Service) IsInstalled() bool {
	return s.configMgr.GetXrayPath() != ""
}

// GetInstallCommand returns command to install Xray
func (s *Service) GetInstallCommand() string {
	arch := runtime.GOARCH
	var xrayArch string

	switch arch {
	case "amd64":
		xrayArch = "64"
	case "arm64":
		xrayArch = "arm64-v8a"
	default:
		xrayArch = "64"
	}

	return fmt.Sprintf(
		"curl -L -o /tmp/xray.zip "https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-%s.zip" && "+
		"unzip -o /tmp/xray.zip -d /tmp/xray && "+
		"sudo mv /tmp/xray/xray /usr/local/bin/ && "+
		"sudo chmod +x /usr/local/bin/xray && "+
		"rm -rf /tmp/xray /tmp/xray.zip",
		xrayArch,
	)
}

// InstallXray downloads and installs Xray
func (s *Service) InstallXray() error {
	cmd := s.GetInstallCommand()
	parts := strings.Split(cmd, " && ")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		args := strings.Fields(part)
		if len(args) == 0 {
			continue
		}

		command := exec.Command(args[0], args[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", part, err)
		}
	}

	return nil
}

// FormatSpeed formats speed in bytes/sec to human readable
func FormatSpeed(bytesPerSec int64) string {
	if bytesPerSec == 0 {
		return "0 B/s"
	}

	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	value := float64(bytesPerSec)
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.1f %s", value, units[unitIndex])
}

// FormatBytes formats bytes to human readable
func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(bytes)
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.1f %s", value, units[unitIndex])
}
