package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/100nandoo/inti/internal/config"
	"github.com/100nandoo/inti/internal/gemini"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	stateInput      appState = iota // user is typing a command
	stateProcessing                 // waiting for Gemini API or playback
	stateMenu                       // interactive command picker is open
)

const (
	headerHeight    = 1
	inputAreaHeight = 1
	statusBarHeight = 1
	menuMaxLines    = 9
)

type logKind string

const (
	kindCommand logKind = "command"
	kindSystem  logKind = "system"
	kindError   logKind = "error"
	kindSuccess logKind = "success"
)

type logEntry struct {
	kind    logKind
	content string
	time    time.Time
}

type commandItem struct {
	name string
	desc string
}

func (c commandItem) FilterValue() string { return c.name }
func (c commandItem) Title() string       { return c.name }
func (c commandItem) Description() string { return c.desc }

type model struct {
	cfg    *config.Config
	gemini *gemini.Client

	input    textinput.Model
	spinner  spinner.Model
	viewport viewport.Model
	cmdList  list.Model

	state        appState
	currentVoice string
	currentModel string
	lastPCM      []byte
	history      []logEntry

	width  int
	height int
}

func newCompactDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetHeight(1)
	d.SetSpacing(0)
	return d
}

func newModel(cfg *config.Config, g *gemini.Client) model {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.Focus()
	ti.CharLimit = 500

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#34C759"))

	vp := viewport.New(0, 0)

	commands := []list.Item{
		commandItem{"speak", "speak <text>   — synthesize and play"},
		commandItem{"voice", "voice <name>   — change voice"},
		commandItem{"model", "model <name>   — change TTS model"},
		commandItem{"export", "export [path]  — save last audio as WAV"},
		commandItem{"pdf", "pdf <file>     — convert PDF pages to images"},
		commandItem{"status", "status         — show configuration"},
		commandItem{"clear", "clear          — clear history"},
		commandItem{"help", "help           — show help"},
		commandItem{"quit", "quit           — exit"},
	}
	cmdList := list.New(commands, newCompactDelegate(), 0, menuMaxLines)
	cmdList.Title = "Commands  ↑↓ navigate · enter select · esc dismiss"
	cmdList.SetShowHelp(false)
	cmdList.SetShowStatusBar(false)
	cmdList.SetFilteringEnabled(false)

	m := model{
		cfg:          cfg,
		gemini:       g,
		input:        ti,
		spinner:      sp,
		viewport:     vp,
		cmdList:      cmdList,
		state:        stateInput,
		currentVoice: cfg.DefaultVoice,
		currentModel: cfg.DefaultModel,
	}

	m.history = []logEntry{
		{kind: kindSystem, content: "INTI v1.0 — Gemini TTS", time: time.Now()},
		{kind: kindSystem, content: fmt.Sprintf("model: gemini-3.1-flash-tts-preview | voice: %s | 24kHz PCM-16", cfg.DefaultVoice), time: time.Now()},
		{kind: kindSystem, content: "press enter to see available commands", time: time.Now()},
	}
	m.viewport.SetContent(m.renderHistoryString())
	m.viewport.GotoBottom()

	return m
}

func Run(cfg *config.Config, g *gemini.Client) error {
	p := tea.NewProgram(newModel(cfg, g), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) viewportHeight(s appState) int {
	fixed := headerHeight + statusBarHeight
	switch s {
	case stateMenu:
		return max(1, m.height-fixed-menuMaxLines)
	default:
		return max(1, m.height-fixed-inputAreaHeight)
	}
}

func (m model) appendLog(kind logKind, content string) model {
	m.history = append(m.history, logEntry{kind: kind, content: content, time: time.Now()})
	m.viewport.SetContent(m.renderHistoryString())
	m.viewport.GotoBottom()
	return m
}

func (m model) renderHistoryString() string {
	lines := make([]string, 0, len(m.history))
	for _, e := range m.history {
		switch e.kind {
		case kindCommand:
			lines = append(lines, promptStyle.Render("❯ ")+cmdStyle.Render(e.content))
		case kindSystem:
			lines = append(lines, dimStyle.Render("  "+e.content))
		case kindError:
			lines = append(lines, errorStyle.Render("✗ "+e.content))
		case kindSuccess:
			lines = append(lines, successStyle.Render("✓ "+e.content))
		}
	}
	return strings.Join(lines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
