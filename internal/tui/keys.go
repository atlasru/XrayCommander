package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/key"
	"github.com/atlasru/xraycommander/pkg/models"
)

// handleKeyMsg handles keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMain:
		return m.handleMainKeys(msg)
	case StateProfileList:
		return m.handleProfileListKeys(msg)
	case StateProfileCreate, StateProfileEdit:
		return m.handleProfileFormKeys(msg)
	case StateLogs:
		return m.handleLogsKeys(msg)
	case StateConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case StateXrayNotFound:
		return m.handleXrayNotFoundKeys(msg)
	}

	return m, nil
}

func (m Model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Start):
		if len(m.profiles) > 0 {
			m.state = StateProfileList
			m.setMessage("Select a profile to connect", "info")
		} else {
			m.setMessage("No profiles available. Create one first.", "warning")
		}

	case key.Matches(msg, m.keys.Stop):
		if err := m.xrayService.Stop(); err != nil {
			m.setMessage(fmt.Sprintf("Failed to stop: %v", err), "error")
		} else {
			m.setMessage("Xray stopped", "success")
		}

	case key.Matches(msg, m.keys.Restart):
		if err := m.xrayService.Restart(); err != nil {
			m.setMessage(fmt.Sprintf("Failed to restart: %v", err), "error")
		} else {
			m.setMessage("Xray restarted", "success")
		}

	case key.Matches(msg, m.keys.Logs):
		m.state = StateLogs

	case msg.String() == "p" || msg.String() == "P":
		m.state = StateProfileList
		m.loadProfiles()
	}

	return m, nil
}

func (m Model) handleProfileListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Esc):
		m.state = StateMain
		m.clearMessage()

	case key.Matches(msg, m.keys.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}

	case key.Matches(msg, m.keys.Down):
		if m.selectedIdx < len(m.profiles)-1 {
			m.selectedIdx++
		}

	case key.Matches(msg, m.keys.Enter):
		if m.selectedIdx < len(m.profiles) {
			profile := m.profiles[m.selectedIdx]
			m.currentProfile = &profile

			if err := m.xrayService.Start(&profile); err != nil {
				m.setMessage(fmt.Sprintf("Connection failed: %v", err), "error")
			} else {
				m.setMessage(fmt.Sprintf("Connected to %s", profile.Name), "success")
				m.state = StateMain
			}
		}

	case msg.String() == "n" || msg.String() == "N":
		m.initProfileForm(nil)
		m.state = StateProfileCreate

	case msg.String() == "e" || msg.String() == "E":
		if m.selectedIdx < len(m.profiles) {
			profile := m.profiles[m.selectedIdx]
			m.initProfileForm(&profile)
			m.state = StateProfileEdit
		}

	case msg.String() == "d" || msg.String() == "D":
		if m.selectedIdx < len(m.profiles) {
			m.state = StateConfirmDelete
		}

	case msg.String() == "t" || msg.String() == "T":
		if m.selectedIdx < len(m.profiles) {
			profile := m.profiles[m.selectedIdx]
			latency, err := m.xrayService.TestConnection(profile.Address, profile.Port)
			if err != nil {
				m.setMessage(fmt.Sprintf("Connection test failed: %v", err), "error")
			} else {
				m.setMessage(fmt.Sprintf("Latency: %d ms", latency), "success")
			}
		}
	}

	return m, nil
}

func (m Model) handleProfileFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Esc):
		m.state = StateProfileList
		m.formInputs = nil

	case key.Matches(msg, m.keys.Save):
		return m.saveProfile()

	case msg.String() == "tab":
		m.formFocus = (m.formFocus + 1) % len(m.formInputs)
		return m.updateFormFocus()

	case msg.String() == "shift+tab":
		m.formFocus--
		if m.formFocus < 0 {
			m.formFocus = len(m.formInputs) - 1
		}
		return m.updateFormFocus()

	default:
		if m.formFocus < len(m.formInputs) {
			var cmd tea.Cmd
			m.formInputs[m.formFocus], cmd = m.formInputs[m.formFocus].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) updateFormFocus() (tea.Model, tea.Cmd) {
	for i := range m.formInputs {
		if i == m.formFocus {
			m.formInputs[i].Focus()
		} else {
			m.formInputs[i].Blur()
		}
	}
	return m, nil
}

func (m Model) saveProfile() (tea.Model, tea.Cmd) {
	if len(m.formInputs) < 4 {
		return m, nil
	}

	port, _ := strconv.Atoi(m.formInputs[2].Value())

	profile := models.Profile{
		ID:          m.editingID,
		Name:        m.formInputs[0].Value(),
		Address:     m.formInputs[1].Value(),
		Port:        port,
		UUID:        m.formInputs[3].Value(),
		Encryption:  m.formInputs[4].Value(),
		Flow:        m.formInputs[5].Value(),
		Network:     m.formInputs[6].Value(),
		Security:    m.formInputs[7].Value(),
		SNI:         m.formInputs[8].Value(),
		Fingerprint: m.formInputs[9].Value(),
		PublicKey:   m.formInputs[10].Value(),
		ShortID:     m.formInputs[11].Value(),
	}

	if profile.Encryption == "" {
		profile.Encryption = "none"
	}
	if profile.Network == "" {
		profile.Network = "tcp"
	}

	if err := profile.Validate(); err != nil {
		m.setMessage(fmt.Sprintf("Validation error: %v", err), "error")
		return m, nil
	}

	if err := m.configMgr.SaveProfile(&profile); err != nil {
		m.setMessage(fmt.Sprintf("Save failed: %v", err), "error")
		return m, nil
	}

	m.loadProfiles()
	m.state = StateProfileList
	m.formInputs = nil
	m.setMessage("Profile saved successfully", "success")

	return m, nil
}

func (m *Model) initProfileForm(profile *models.Profile) {
	m.formInputs = make([]textinput.Model, 12)

	fields := []struct {
		placeholder string
		value       string
		charLimit   int
	}{
		{"Profile Name", "", 50},
		{"Server Address", "", 255},
		{"Port", "443", 5},
		{"UUID", "", 36},
		{"Encryption (none)", "none", 10},
		{"Flow (xtls-rprx-vision)", "", 30},
		{"Network (tcp)", "tcp", 10},
		{"Security (tls/reality)", "", 10},
		{"SNI", "", 255},
		{"Fingerprint (chrome)", "", 20},
		{"Public Key (Reality)", "", 100},
		{"Short ID (Reality)", "", 20},
	}

	if profile != nil {
		fields[0].value = profile.Name
		fields[1].value = profile.Address
		fields[2].value = strconv.Itoa(profile.Port)
		fields[3].value = profile.UUID
		fields[4].value = profile.Encryption
		fields[5].value = profile.Flow
		fields[6].value = profile.Network
		fields[7].value = profile.Security
		fields[8].value = profile.SNI
		fields[9].value = profile.Fingerprint
		fields[10].value = profile.PublicKey
		fields[11].value = profile.ShortID
		m.editingID = profile.ID
	} else {
		m.editingID = ""
	}

	for i, field := range fields {
		input := textinput.New()
		input.Placeholder = field.placeholder
		input.SetValue(field.value)
		input.CharLimit = field.charLimit
		input.Width = 40
		m.formInputs[i] = input
	}

	m.formFocus = 0
	m.formInputs[0].Focus()
}

func (m Model) handleLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Esc) {
		m.state = StateMain
	}
	return m, nil
}

func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.selectedIdx < len(m.profiles) {
			profile := m.profiles[m.selectedIdx]
			if err := m.configMgr.DeleteProfile(profile.ID); err != nil {
				m.setMessage(fmt.Sprintf("Delete failed: %v", err), "error")
			} else {
				m.setMessage("Profile deleted", "success")
				m.loadProfiles()
				if m.selectedIdx >= len(m.profiles) && m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}
		}
		m.state = StateProfileList

	case "n", "N", "esc":
		m.state = StateProfileList
	}

	return m, nil
}

func (m Model) handleXrayNotFoundKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i", "I":
		m.setMessage("Installing Xray...", "info")
		if err := m.xrayService.InstallXray(); err != nil {
			m.setMessage(fmt.Sprintf("Installation failed: %v", err), "error")
		} else {
			m.setMessage("Xray installed successfully!", "success")
			m.state = StateMain
		}

	case "q", "Q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}
