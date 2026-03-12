package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/atlasru/xraycommander/internal/config"
	"github.com/atlasru/xraycommander/internal/xray"
	"github.com/atlasru/xraycommander/pkg/models"
)

// ViewState represents the current view state
type ViewState int

const (
	StateMain ViewState = iota
	StateProfileList
	StateProfileEdit
	StateProfileCreate
	StateLogs
	StateConfirmDelete
	StateXrayNotFound
	StateSpeedTest
)

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Enter   key.Binding
	Esc     key.Binding
	Quit    key.Binding
	Save    key.Binding
	Delete  key.Binding
	Start   key.Binding
	Stop    key.Binding
	Restart key.Binding
	Logs    key.Binding
	Test    key.Binding
	Help    key.Binding
}

// DefaultKeyMap returns default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Delete: key.NewBinding(
			key.WithKeys("ctrl+d", "delete"),
			key.WithHelp("ctrl+d", "delete"),
		),
		Start: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "start"),
		),
		Stop: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "stop"),
		),
		Restart: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "restart"),
		),
		Logs: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "logs"),
		),
		Test: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "test"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// Styles
type Styles struct {
	AppStyle          lipgloss.Style
	TitleStyle        lipgloss.Style
	StatusBarStyle    lipgloss.Style
	ConnectedStyle    lipgloss.Style
	DisconnectedStyle lipgloss.Style
	ErrorStyle        lipgloss.Style
	SuccessStyle      lipgloss.Style
	WarningStyle      lipgloss.Style
	HelpStyle         lipgloss.Style
	BorderStyle       lipgloss.Style
	SelectedStyle     lipgloss.Style
	NormalStyle       lipgloss.Style
	LabelStyle        lipgloss.Style
	ValueStyle        lipgloss.Style
	HeaderStyle       lipgloss.Style
}

// DefaultStyles returns default styles
func DefaultStyles() Styles {
	return Styles{
		AppStyle: lipgloss.NewStyle().
			Padding(1, 2).
			Background(lipgloss.Color("#1a1b26")),

		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7aa2f7")).
			Background(lipgloss.Color("#1a1b26")).
			Padding(0, 1).
			MarginBottom(1),

		StatusBarStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#24283b")).
			Foreground(lipgloss.Color("#a9b1d6")).
			Padding(0, 1),

		ConnectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a")),

		DisconnectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f7768e")),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f7768e")),

		SuccessStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a")),

		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0af68")),

		HelpStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#24283b")).
			Foreground(lipgloss.Color("#565f89")),

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#414868")),

		SelectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#283457")).
			Foreground(lipgloss.Color("#7aa2f7")),

		NormalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5")),

		LabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7aa2f7")),

		ValueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5")),

		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bb9af7")),
	}
}

// Model is the main TUI model
type Model struct {
	state       ViewState
	keys        KeyMap
	help        help.Model
	styles      Styles
	width       int
	height      int

	// Managers
	configMgr   *config.Manager
	xrayService *xray.Service

	// Data
	profiles    []models.Profile
	selectedIdx int
	currentProfile *models.Profile

	// Profile form
	formInputs  []textinput.Model
	formFocus   int
	editingID   string

	// Lists
	profileList list.Model

	// Logs
	logsTable   table.Model

	// Messages
	message     string
	messageType string // "error", "success", "warning"
}

// NewModel creates a new TUI model
func NewModel() Model {
	// Initialize config manager
	configMgr, err := config.NewManager()
	if err != nil {
		// Handle error gracefully
		configMgr = nil
	}

	m := Model{
		state:      StateMain,
		keys:       DefaultKeyMap(),
		help:       help.New(),
		styles:     DefaultStyles(),
		configMgr:  configMgr,
		formInputs: make([]textinput.Model, 0),
	}

	// Initialize Xray service
	if configMgr != nil {
		m.xrayService = xray.NewService(configMgr, func() {
			// Trigger update when service updates
		})
	}

	// Load profiles
	m.loadProfiles()

	// Check if Xray is installed
	if configMgr != nil && !m.xrayService.IsInstalled() {
		m.state = StateXrayNotFound
	}

	return m
}

func (m *Model) loadProfiles() {
	if m.configMgr == nil {
		return
	}

	profiles, err := m.configMgr.GetProfiles()
	if err != nil {
		m.profiles = []models.Profile{}
		return
	}
	m.profiles = profiles
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tickMsg:
		// Periodic update
		return m, tickCmd()
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content string

	switch m.state {
	case StateMain:
		content = m.viewMain()
	case StateProfileList:
		content = m.viewProfileList()
	case StateProfileEdit, StateProfileCreate:
		content = m.viewProfileForm()
	case StateLogs:
		content = m.viewLogs()
	case StateConfirmDelete:
		content = m.viewConfirmDelete()
	case StateXrayNotFound:
		content = m.viewXrayNotFound()
	default:
		content = m.viewMain()
	}

	// Add status bar
	statusBar := m.renderStatusBar()

	// Add help
	helpView := m.renderHelp()

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusBar,
		helpView,
	)
}

func (m Model) renderStatusBar() string {
	status := m.xrayService.GetStatus()

	var statusText string
	if status.Connected {
		statusText = m.styles.ConnectedStyle.Render("● Connected")
	} else {
		statusText = m.styles.DisconnectedStyle.Render("○ Disconnected")
	}

	var profileText string
	if status.CurrentProfile != nil {
		profileText = fmt.Sprintf(" | Profile: %s", status.CurrentProfile.Name)
	}

	var speedText string
	if status.Connected {
		speedText = fmt.Sprintf(" | ↑ %s ↓ %s", 
			xray.FormatSpeed(status.UploadSpeed),
			xray.FormatSpeed(status.DownloadSpeed))
	}

	content := fmt.Sprintf("%s%s%s", statusText, profileText, speedText)

	if m.message != "" {
		var msgStyle lipgloss.Style
		switch m.messageType {
		case "error":
			msgStyle = m.styles.ErrorStyle
		case "success":
			msgStyle = m.styles.SuccessStyle
		case "warning":
			msgStyle = m.styles.WarningStyle
		default:
			msgStyle = m.styles.NormalStyle
		}
		content = fmt.Sprintf("%s | %s", content, msgStyle.Render(m.message))
	}

	return m.styles.StatusBarStyle.
		Width(m.width - 2).
		Render(content)
}

func (m Model) renderHelp() string {
	bindings := []key.Binding{
		m.keys.Up,
		m.keys.Down,
		m.keys.Enter,
		m.keys.Esc,
		m.keys.Quit,
	}

	if m.state == StateMain {
		bindings = append(bindings, 
			m.keys.Start, 
			m.keys.Stop, 
			m.keys.Restart,
			m.keys.Logs,
		)
	}

	if m.state == StateProfileList {
		bindings = append(bindings,
			m.keys.Save,
			m.keys.Delete,
		)
	}

	return m.styles.HelpStyle.
		Width(m.width - 2).
		Render(m.help.ShortHelpView(bindings))
}

// tickCmd creates a tick command for periodic updates
type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(1, func(t interface{}) tea.Msg {
		return tickMsg{}
	})
}
