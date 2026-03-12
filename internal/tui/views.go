package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/atlasru/xraycommander/internal/xray"
)

// viewMain renders the main dashboard
func (m Model) viewMain() string {
	var b strings.Builder

	// Title
	b.WriteString(m.styles.TitleStyle.Render("XrayCommander"))
	b.WriteString("

")

	// Status card
	status := m.xrayService.GetStatus()

	var statusCard string
	if status.Connected {
		statusCard = m.renderConnectedCard(status)
	} else {
		statusCard = m.renderDisconnectedCard()
	}

	b.WriteString(m.styles.BorderStyle.Render(statusCard))
	b.WriteString("

")

	// Quick actions
	b.WriteString(m.styles.HeaderStyle.Render("Quick Actions"))
	b.WriteString("
")

	actions := []string{
		"[Ctrl+O] Start Connection",
		"[Ctrl+X] Stop Connection",
		"[Ctrl+R] Restart Service",
		"[Ctrl+L] View Logs",
		"[P] Manage Profiles",
		"[Q] Quit",
	}

	for _, action := range actions {
		b.WriteString(fmt.Sprintf("  %s
", m.styles.NormalStyle.Render(action)))
	}

	// Recent profiles
	if len(m.profiles) > 0 {
		b.WriteString("
")
		b.WriteString(m.styles.HeaderStyle.Render("Recent Profiles"))
		b.WriteString("
")

		for i, profile := range m.profiles {
			if i >= 5 {
				break
			}

			marker := "  "
			if status.CurrentProfile != nil && status.CurrentProfile.ID == profile.ID {
				marker = m.styles.ConnectedStyle.Render("→ ")
			}

			b.WriteString(fmt.Sprintf("%s%s (%s:%d)
", 
				marker,
				m.styles.NormalStyle.Render(profile.Name),
				profile.Address, 
				profile.Port))
		}
	}

	return b.String()
}

func (m Model) renderConnectedCard(status xray.ConnectionStatus) string {
	var b strings.Builder

	b.WriteString(m.styles.ConnectedStyle.Render("● CONNECTED"))
	b.WriteString("

")

	if status.CurrentProfile != nil {
		b.WriteString(fmt.Sprintf("Profile: %s
", m.styles.ValueStyle.Render(status.CurrentProfile.Name)))
		b.WriteString(fmt.Sprintf("Server: %s:%d
", 
			status.CurrentProfile.Address, 
			status.CurrentProfile.Port))
	}

	b.WriteString(fmt.Sprintf("Upload: %s (Total: %s)
", 
		xray.FormatSpeed(status.UploadSpeed),
		xray.FormatBytes(status.TotalUpload)))
	b.WriteString(fmt.Sprintf("Download: %s (Total: %s)
",
		xray.FormatSpeed(status.DownloadSpeed),
		xray.FormatBytes(status.TotalDownload)))

	if status.Latency > 0 {
		b.WriteString(fmt.Sprintf("Latency: %d ms
", status.Latency))
	}

	return b.String()
}

func (m Model) renderDisconnectedCard() string {
	return fmt.Sprintf("%s

%s
%s",
		m.styles.DisconnectedStyle.Render("○ DISCONNECTED"),
		"Xray service is not running.",
		"Select a profile and press Ctrl+O to connect.")
}

// viewProfileList renders the profile list
func (m Model) viewProfileList() string {
	var b strings.Builder

	b.WriteString(m.styles.TitleStyle.Render("Profile Management"))
	b.WriteString("

")

	if len(m.profiles) == 0 {
		b.WriteString(m.styles.WarningStyle.Render("No profiles found."))
		b.WriteString("

")
		b.WriteString("Press [N] to create a new profile or [Ctrl+Q] to go back.")
		return b.String()
	}

	// Profile table
	rows := make([][]string, len(m.profiles))
	for i, profile := range m.profiles {
		security := profile.Security
		if security == "" {
			security = "none"
		}

		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			profile.Name,
			fmt.Sprintf("%s:%d", profile.Address, profile.Port),
			profile.Network,
			security,
		}
	}

	// Render table manually
	b.WriteString(fmt.Sprintf("%-4s %-20s %-25s %-10s %-10s
", "#", "Name", "Server", "Network", "Security"))
	b.WriteString(strings.Repeat("─", 75))
	b.WriteString("
")

	for i, row := range rows {
		line := fmt.Sprintf("%-4s %-20s %-25s %-10s %-10s", row[0], row[1], row[2], row[3], row[4])
		if i == m.selectedIdx {
			line = m.styles.SelectedStyle.Render(line)
		} else {
			line = m.styles.NormalStyle.Render(line)
		}
		b.WriteString(line + "
")
	}

	b.WriteString("
")
	b.WriteString("[Enter] Connect/Edit  [N] New  [E] Edit  [D] Delete  [T] Test  [Esc] Back")

	return b.String()
}

// viewProfileForm renders the profile creation/edit form
func (m Model) viewProfileForm() string {
	var b strings.Builder

	if m.state == StateProfileCreate {
		b.WriteString(m.styles.TitleStyle.Render("Create New Profile"))
	} else {
		b.WriteString(m.styles.TitleStyle.Render("Edit Profile"))
	}
	b.WriteString("

")

	labels := []string{
		"Name:        ",
		"Address:     ",
		"Port:        ",
		"UUID:        ",
		"Encryption:  ",
		"Flow:        ",
		"Network:     ",
		"Security:    ",
		"SNI:         ",
		"Fingerprint: ",
		"Public Key:  ",
		"Short ID:    ",
	}

	for i, input := range m.formInputs {
		if i < len(labels) {
			label := m.styles.LabelStyle.Render(labels[i])
			value := input.View()

			if i == m.formFocus {
				value = m.styles.SelectedStyle.Render(value)
			}

			b.WriteString(fmt.Sprintf("%s%s
", label, value))
		}
	}

	b.WriteString("
")
	b.WriteString("[Tab] Next Field  [Shift+Tab] Prev  [Ctrl+S] Save  [Esc] Cancel")

	return b.String()
}

// viewLogs renders the logs view
func (m Model) viewLogs() string {
	var b strings.Builder

	b.WriteString(m.styles.TitleStyle.Render("Xray Logs"))
	b.WriteString("

")

	logs := m.xrayService.GetLogs()
	if len(logs) == 0 {
		b.WriteString(m.styles.NormalStyle.Render("No logs available. Start Xray to see logs."))
	} else {
		// Show last 20 lines
		start := 0
		if len(logs) > 20 {
			start = len(logs) - 20
		}

		for i := start; i < len(logs); i++ {
			line := logs[i]
			// Colorize log levels
			if strings.Contains(line, "ERROR") || strings.Contains(line, "error") {
				line = m.styles.ErrorStyle.Render(line)
			} else if strings.Contains(line, "WARN") || strings.Contains(line, "warning") {
				line = m.styles.WarningStyle.Render(line)
			} else {
				line = m.styles.NormalStyle.Render(line)
			}
			b.WriteString(line + "
")
		}
	}

	b.WriteString("
")
	b.WriteString("[Esc] Back  [Ctrl+C] Clear")

	return b.String()
}

// viewConfirmDelete renders the delete confirmation
func (m Model) viewConfirmDelete() string {
	var b strings.Builder

	b.WriteString(m.styles.TitleStyle.Render("Confirm Delete"))
	b.WriteString("

")

	if m.selectedIdx < len(m.profiles) {
		profile := m.profiles[m.selectedIdx]
		b.WriteString(fmt.Sprintf("Are you sure you want to delete profile '%s'?

", 
			m.styles.WarningStyle.Render(profile.Name)))
		b.WriteString(fmt.Sprintf("Server: %s:%d

", profile.Address, profile.Port))
	}

	b.WriteString("[Y] Yes, delete  [N] No, cancel")

	return b.String()
}

// viewXrayNotFound renders the Xray installation prompt
func (m Model) viewXrayNotFound() string {
	var b strings.Builder

	b.WriteString(m.styles.TitleStyle.Render("Xray Not Found"))
	b.WriteString("

")

	b.WriteString(m.styles.WarningStyle.Render("Xray-core binary not found on your system."))
	b.WriteString("

")

	b.WriteString("XrayCommander requires Xray-core to function.
")
	b.WriteString("You can install it automatically or manually.

")

	b.WriteString(m.styles.HeaderStyle.Render("Automatic Installation"))
	b.WriteString("
")
	b.WriteString("Press [I] to install Xray-core automatically.
")
	b.WriteString("This will download and install the latest version.

")

	b.WriteString(m.styles.HeaderStyle.Render("Manual Installation"))
	b.WriteString("
")
	b.WriteString("Run the following command:
")
	b.WriteString(m.styles.NormalStyle.Render(m.xrayService.GetInstallCommand()))
	b.WriteString("

")

	b.WriteString("[I] Install  [Q] Quit")

	return b.String()
}

// setMessage sets a temporary message
func (m *Model) setMessage(msg string, msgType string) {
	m.message = msg
	m.messageType = msgType
}

// clearMessage clears the message
func (m *Model) clearMessage() {
	m.message = ""
	m.messageType = ""
}
