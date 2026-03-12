package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Profile represents a VLESS connection profile
type Profile struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Address     string    `json:"address" yaml:"address"`
	Port        int       `json:"port" yaml:"port"`
	UUID        string    `json:"uuid" yaml:"uuid"` // VLESS UUID
	Encryption  string    `json:"encryption" yaml:"encryption"` // Usually "none"
	Flow        string    `json:"flow,omitempty" yaml:"flow,omitempty"` // xtls-rprx-vision
	Network     string    `json:"network" yaml:"network"` // tcp, ws, grpc
	Security    string    `json:"security" yaml:"security"` // tls, xtls, reality, none
	SNI         string    `json:"sni,omitempty" yaml:"sni,omitempty"`
	ALPN        []string  `json:"alpn,omitempty" yaml:"alpn,omitempty"`
	Fingerprint string    `json:"fp,omitempty" yaml:"fp,omitempty"` // Chrome, Firefox, etc
	PublicKey   string    `json:"pbk,omitempty" yaml:"pbk,omitempty"` // Reality
	ShortID     string    `json:"sid,omitempty" yaml:"sid,omitempty"` // Reality
	SpiderX     string    `json:"spx,omitempty" yaml:"spx,omitempty"` // Reality
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

// Validate checks if the profile has all required fields
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	if p.Address == "" {
		return fmt.Errorf("server address is required")
	}
	if p.Port <= 0 || p.Port > 65535 {
		return fmt.Errorf("invalid port number")
	}
	if p.UUID == "" {
		return fmt.Errorf("UUID is required")
	}
	if p.Encryption == "" {
		p.Encryption = "none"
	}
	if p.Network == "" {
		p.Network = "tcp"
	}
	return nil
}

// ToXrayConfig converts profile to Xray JSON config
func (p *Profile) ToXrayConfig() map[string]interface{} {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"access":   "",
			"error":    "",
			"loglevel": "warning",
		},
		"inbounds": []map[string]interface{}{
			{
				"tag":      "socks",
				"port":     10808,
				"listen":   "127.0.0.1",
				"protocol": "socks",
				"settings": map[string]interface{}{
					"auth": "noauth",
					"udp":  true,
					"ip":   "127.0.0.1",
				},
				"sniffing": map[string]interface{}{
					"enabled":      true,
					"destOverride": []string{"http", "tls"},
				},
			},
			{
				"tag":      "http",
				"port":     10809,
				"listen":   "127.0.0.1",
				"protocol": "http",
				"settings": map[string]interface{}{},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"tag":      "proxy",
				"protocol": "vless",
				"settings": map[string]interface{}{
					"vnext": []map[string]interface{}{
						{
							"address": p.Address,
							"port":    p.Port,
							"users": []map[string]interface{}{
								{
									"id":         p.UUID,
									"encryption": p.Encryption,
									"flow":       p.Flow,
								},
							},
						},
					},
				},
				"streamSettings": p.buildStreamSettings(),
			},
			{
				"tag":      "direct",
				"protocol": "freedom",
			},
			{
				"tag":      "block",
				"protocol": "blackhole",
				"settings": map[string]interface{}{
					"response": map[string]interface{}{
						"type": "http",
					},
				},
			},
		},
		"routing": map[string]interface{}{
			"domainStrategy": "IPIfNonMatch",
			"rules": []map[string]interface{}{
				{
					"type":        "field",
					"ip":          []string{"geoip:private"},
					"outboundTag": "direct",
				},
				{
					"type":        "field",
					"domain":      []string{"geosite:private"},
					"outboundTag": "direct",
				},
			},
		},
	}
	return config
}

func (p *Profile) buildStreamSettings() map[string]interface{} {
	settings := map[string]interface{}{
		"network":  p.Network,
		"security": p.Security,
	}

	if p.Security == "tls" || p.Security == "xtls" {
		tlsSettings := map[string]interface{}{
			"serverName":    p.SNI,
			"allowInsecure": false,
		}
		if len(p.ALPN) > 0 {
			tlsSettings["alpn"] = p.ALPN
		}
		if p.Fingerprint != "" {
			tlsSettings["fingerprint"] = p.Fingerprint
		}
		if p.Security == "tls" {
			settings["tlsSettings"] = tlsSettings
		} else {
			settings["xtlsSettings"] = tlsSettings
		}
	}

	if p.Security == "reality" {
		realitySettings := map[string]interface{}{
			"serverName":  p.SNI,
			"fingerprint": p.Fingerprint,
			"show":        false,
			"publicKey":   p.PublicKey,
			"shortId":     p.ShortID,
			"spiderX":     p.SpiderX,
		}
		settings["realitySettings"] = realitySettings
	}

	return settings
}

// ToJSON exports profile as JSON
func (p *Profile) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ConnectionStatus represents current connection state
type ConnectionStatus struct {
	Connected      bool       `json:"connected"`
	CurrentProfile *Profile   `json:"current_profile,omitempty"`
	UploadSpeed    int64      `json:"upload_speed"`   // bytes/sec
	DownloadSpeed  int64      `json:"download_speed"` // bytes/sec
	TotalUpload    int64      `json:"total_upload"`
	TotalDownload  int64      `json:"total_download"`
	Latency        int        `json:"latency_ms"`
	StartedAt      time.Time  `json:"started_at,omitempty"`
	PID            int        `json:"pid,omitempty"`
}

// Stats represents traffic statistics
type Stats struct {
	Timestamp     time.Time `json:"timestamp"`
	UploadBytes   int64     `json:"upload_bytes"`
	DownloadBytes int64     `json:"download_bytes"`
}
