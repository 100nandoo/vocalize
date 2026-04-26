package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/100nandoo/inti/internal/audio"
	"github.com/100nandoo/inti/internal/config"
	"github.com/100nandoo/inti/internal/gemini"
	"github.com/100nandoo/inti/internal/pdf"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch m.state {
		case stateInput:
			return m.updateInput(msg)
		case stateMenu:
			return m.updateMenu(msg)
		case stateProcessing:
			return m, nil
		}

	case speechReadyMsg:
		m = m.appendLog(kindSuccess, fmt.Sprintf("synthesized with voice %s — playing...", msg.voice))
		m.lastPCM = msg.pcm
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, playAudioCmd(msg.pcm)

	case speechErrMsg:
		if gemini.IsRateLimit(msg.err) {
			m = m.appendLog(kindError, "rate limited — wait a moment and try again")
		} else {
			m = m.appendLog(kindError, fmt.Sprintf("error: %v", msg.err))
		}
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case playbackDoneMsg:
		m = m.appendLog(kindSuccess, "playback complete")
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case exportDoneMsg:
		m = m.appendLog(kindSuccess, fmt.Sprintf("saved to %s", msg.path))
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case pdfDoneMsg:
		m = m.appendLog(kindSuccess, fmt.Sprintf("converted %d pages → %s/", msg.pages, msg.outputDir))
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case pdfErrMsg:
		m = m.appendLog(kindError, fmt.Sprintf("pdf error: %v", msg.err))
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.state == stateInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.viewport.Width = msg.Width
	m.viewport.Height = m.viewportHeight(m.state)
	m.cmdList.SetSize(msg.Width, menuMaxLines)
	return m, nil
}

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, tea.Quit

	case tea.KeyEnter:
		raw := strings.TrimSpace(m.input.Value())
		if raw == "" {
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		m.input.SetValue("")
		return m.handleCommand(raw)

	case tea.KeyUp, tea.KeyDown:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateInput
		m.viewport.Height = m.viewportHeight(stateInput)
		m.input.Focus()
		return m, textinput.Blink

	case tea.KeyEnter:
		selected := m.cmdList.SelectedItem()
		if selected == nil {
			return m, nil
		}
		ci := selected.(commandItem)
		m.state = stateInput
		m.viewport.Height = m.viewportHeight(stateInput)
		m.input.Focus()
		m.input.SetValue(ci.name + " ")
		m.input.SetCursor(len(ci.name) + 1)
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	m.cmdList, cmd = m.cmdList.Update(msg)
	return m, cmd
}

func (m model) handleCommand(raw string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return m, nil
	}

	m = m.appendLog(kindCommand, raw)
	verb := strings.ToLower(parts[0])
	args := strings.Join(parts[1:], " ")

	switch verb {
	case "speak":
		if args == "" {
			m = m.appendLog(kindError, "usage: speak <text>")
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		m.state = stateProcessing
		m = m.appendLog(kindSystem, fmt.Sprintf("synthesizing: %q", args))
		return m, synthesizeCmd(m.gemini, args, m.currentVoice, m.currentModel)

	case "voice":
		if args == "" {
			m = m.appendLog(kindError, fmt.Sprintf("usage: voice <name> — available: %v", config.ValidVoices()))
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		name := strings.ToUpper(args[:1]) + strings.ToLower(args[1:])
		if !config.IsValidVoice(name) {
			m = m.appendLog(kindError, fmt.Sprintf("unknown voice %q — available: %v", args, config.ValidVoices()))
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		m.currentVoice = name
		m = m.appendLog(kindSuccess, fmt.Sprintf("voice set to %s", name))
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case "model":
		if args == "" {
			m = m.appendLog(kindError, fmt.Sprintf("usage: model <name> — available: %v", config.ValidModels()))
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		if !config.IsValidModel(args) {
			m = m.appendLog(kindError, fmt.Sprintf("unknown model %q — available: %v", args, config.ValidModels()))
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		m.currentModel = args
		m = m.appendLog(kindSuccess, fmt.Sprintf("model set to %s", args))
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case "export":
		if len(m.lastPCM) == 0 {
			m = m.appendLog(kindError, "nothing to export — run speak first")
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		path := fmt.Sprintf("inti-%d.wav", time.Now().Unix())
		if args != "" {
			path = args
		}
		return m, exportCmd(m.lastPCM, path)

	case "pdf":
		if args == "" {
			m = m.appendLog(kindError, "usage: pdf <file.pdf> [output-dir]")
			m.state = stateMenu
			m.viewport.Height = m.viewportHeight(stateMenu)
			return m, nil
		}
		pdfParts := strings.Fields(args)
		inputPath := pdfParts[0]
		outputDir := ""
		if len(pdfParts) > 1 {
			outputDir = pdfParts[1]
		} else {
			base := filepath.Base(inputPath)
			outputDir = strings.TrimSuffix(base, filepath.Ext(base))
		}
		m.state = stateProcessing
		m = m.appendLog(kindSystem, fmt.Sprintf("converting %s → %s/", inputPath, outputDir))
		return m, pdfConvertCmd(inputPath, outputDir)

	case "status":
		m = m.appendLog(kindSystem, fmt.Sprintf("model:   %s", m.currentModel))
		m = m.appendLog(kindSystem, fmt.Sprintf("voice:   %s", m.currentVoice))
		m = m.appendLog(kindSystem, "audio:   24000 Hz, PCM-16, mono")
		if len(m.lastPCM) > 0 {
			m = m.appendLog(kindSystem, fmt.Sprintf("buffer:  %d bytes (%.1fs)", len(m.lastPCM), float64(len(m.lastPCM))/48000.0))
		}
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case "clear":
		m.history = nil
		m.viewport.SetContent("")
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case "help":
		m = m.appendLog(kindSystem, "speak <text>      — synthesize and play")
		m = m.appendLog(kindSystem, "voice <name>      — set voice")
		m = m.appendLog(kindSystem, "export [path]     — save last audio as WAV")
		m = m.appendLog(kindSystem, "pdf <file> [dir]  — convert PDF pages to images")
		m = m.appendLog(kindSystem, "status            — show configuration")
		m = m.appendLog(kindSystem, "clear             — clear history")
		m = m.appendLog(kindSystem, "help              — show this message")
		m = m.appendLog(kindSystem, "q / ctrl+c        — quit")
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil

	case "q", "quit", "exit":
		return m, tea.Quit

	default:
		m = m.appendLog(kindError, fmt.Sprintf("unknown command %q — press enter for commands", verb))
		m.state = stateMenu
		m.viewport.Height = m.viewportHeight(stateMenu)
		return m, nil
	}
}

func synthesizeCmd(g interface {
	GenerateSpeech(ctx context.Context, text, voice, model string) ([]byte, error)
}, text, voice, model string) tea.Cmd {
	return func() tea.Msg {
		pcm, err := g.GenerateSpeech(context.Background(), text, voice, model)
		if err != nil {
			return speechErrMsg{err: err}
		}
		return speechReadyMsg{pcm: pcm, voice: voice}
	}
}

func playAudioCmd(pcm []byte) tea.Cmd {
	return func() tea.Msg {
		if err := audio.Play(pcm); err != nil {
			return speechErrMsg{err: err}
		}
		return playbackDoneMsg{}
	}
}

func pdfConvertCmd(inputPath, outputDir string) tea.Cmd {
	return func() tea.Msg {
		n, err := pdf.Convert(inputPath, outputDir)
		if err != nil {
			return pdfErrMsg{err: err}
		}
		return pdfDoneMsg{outputDir: outputDir, pages: n}
	}
}

func exportCmd(pcm []byte, path string) tea.Cmd {
	return func() tea.Msg {
		if err := audio.WriteOpusFile(path, pcm, 24000); err != nil {
			return speechErrMsg{err: err}
		}
		return exportDoneMsg{path: path}
	}
}
