package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jhonVitor-rs/devsetup/app/libs"
	"github.com/jhonVitor-rs/devsetup/app/system"
	"github.com/jhonVitor-rs/devsetup/app/ui"
)

type JobState int

const (
	JobPending JobState = iota
	JobRunning
	JobFailed
)

type Job struct {
	Lib    libs.LibStatus
	State  JobState
	Err    error
	Output string
}

type InstallRequestedMsg struct{ Lib libs.LibStatus }

type installDoneMsg struct {
	ID     string
	Err    error
	Output string
	Status libs.LibStatus
}

type sudoCheckedMsg struct {
	Pass string
	OK   bool
}

type App struct {
	sys *system.System

	list    list.Model
	help    help.Model
	keys    *KeyMap
	pwInput textinput.Model

	currentJob     *Job
	showSystemInfo bool
	askPassword    bool
	pwErr          string
	sudoPass       string
	width, height  int
}

func New(sys *system.System, statuses []libs.LibStatus) *App {
	keys := NewKeyMap()

	ti := textinput.New()
	ti.Placeholder = "senha do sudo"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	a := &App{
		sys:     sys,
		keys:    keys,
		help:    help.New(),
		pwInput: ti,
	}

	items := make([]list.Item, len(statuses))
	for i, s := range statuses {
		items[i] = s
	}

	a.list = list.New(items, NewDelegate(keys, a.jobState), 0, 0)
	a.list.Title = "DevSetup"
	a.list.SetShowHelp(false)
	a.list.KeyMap.Quit = keys.quit
	return a
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.help.Width = msg.Width

	case tea.KeyMsg:
		if a.askPassword {
			cmd = a.updatePassword(msg)
			a.resize()
			return a, cmd
		}
		if a.list.FilterState() == list.Filtering {
			break
		}

		if a.showSystemInfo {
			switch {
			case key.Matches(msg, a.keys.openConfig), key.Matches(msg, a.keys.quit):
				a.showSystemInfo = false
			case msg.String() == "ctrl+c":
				return a, tea.Quit
			}
			return a, nil
		}

		switch {
		case a.busy() && key.Matches(msg, a.keys.quit):
			return a, nil
		case key.Matches(msg, a.keys.openHelp):
			a.help.ShowAll = !a.help.ShowAll
			a.resize()
			return a, nil
		case key.Matches(msg, a.keys.openConfig):
			a.showSystemInfo = true
			return a, nil
		}

	case InstallRequestedMsg:
		if a.busy() {
			return a, nil
		}
		a.currentJob = &Job{Lib: msg.Lib, State: JobPending}
		cmds = append(cmds, a.start())

	case sudoCheckedMsg:
		if !msg.OK {
			a.pwErr = "senha incorreta, tente de novo"
			a.pwInput.Reset()
			break
		}
		a.sudoPass, a.askPassword, a.pwErr = msg.Pass, false, ""
		a.pwInput.Blur()
		if a.currentJob != nil && a.currentJob.State == JobPending {
			cmds = append(cmds, a.start())
		}

	case installDoneMsg:
		job := a.currentJob
		if job == nil || job.Lib.Lib.ID != msg.ID {
			break
		}
		cmds = append(cmds, a.updateStatus(msg.Status))

		switch {
		case msg.Err != nil:
			job.State, job.Err, job.Output = JobFailed, msg.Err, tail(msg.Output, 6)
		case !msg.Status.Installed:
			job.State, job.Output = JobFailed, tail(msg.Output, 6)
			job.Err = errors.New("instalação concluída, mas o check não encontrou a lib (talvez seja o PATH; reabra o terminal)")
		default:
			a.currentJob = nil
		}
	}

	a.list, cmd = a.list.Update(msg)
	cmds = append(cmds, cmd)

	if a.askPassword {
		a.pwInput, cmd = a.pwInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	a.resize()
	return a, tea.Batch(cmds...)
}

func (a *App) updatePassword(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		a.askPassword = false
		a.pwInput.Blur()
		a.currentJob = nil
		return nil
	case "enter":
		pw, sys := a.pwInput.Value(), a.sys
		return func() tea.Msg { return sudoCheckedMsg{Pass: pw, OK: sys.CheckSudo(pw)} }
	}

	var cmd tea.Cmd
	a.pwInput, cmd = a.pwInput.Update(msg)
	return cmd
}

func (a *App) start() tea.Cmd {
	job := a.currentJob
	if a.sys.NeedsSudo(job.Lib.Lib) && a.sudoPass == "" {
		a.askPassword, a.pwErr = true, ""
		a.pwInput.Reset()
		return a.pwInput.Focus()
	}
	job.State = JobRunning
	return a.install(job.Lib.Lib, a.sudoPass)
}

func (a *App) install(lib libs.Lib, pass string) tea.Cmd {
	sys := a.sys
	return func() tea.Msg {
		out, err := sys.RunInstall(lib, pass)
		return installDoneMsg{ID: lib.ID, Err: err, Output: out, Status: sys.RunCheck(lib)}
	}
}

func (a *App) busy() bool {
	j := a.currentJob
	return j != nil && (j.State == JobPending || j.State == JobRunning)
}

func (a *App) jobState(id string) (JobState, bool) {
	if j := a.currentJob; j != nil && j.Lib.Lib.ID == id {
		return j.State, true
	}
	return 0, false
}

func (a *App) updateStatus(st libs.LibStatus) tea.Cmd {
	for i, it := range a.list.Items() {
		if ls, ok := it.(libs.LibStatus); ok && ls.Lib.ID == st.Lib.ID {
			return a.list.SetItem(i, st)
		}
	}
	return nil
}

func (a *App) View() string {
	var body string
	switch {
	case a.askPassword:
		body = a.renderPassword()
	case a.showSystemInfo:
		body = a.renderSystemInfo()
	default:
		body = a.list.View()
	}

	if s := a.statusView(); s != "" {
		body += "\n" + s
	}
	return body + "\n" + a.help.View(a.keys)
}

func (a *App) resize() {
	reserved := lipgloss.Height(a.help.View(a.keys)) + 1
	if s := a.statusView(); s != "" {
		reserved += lipgloss.Height(s) + 1
	}
	a.list.SetSize(a.width, max(a.height-reserved, 1))
}

func (a *App) statusView() string {
	j := a.currentJob
	if j == nil {
		return ""
	}
	switch j.State {
	case JobRunning:
		return ui.WarnStyle.Render("… instalando " + j.Lib.Lib.Name)
	case JobFailed:
		s := ui.ErrorStyle.Render(fmt.Sprintf("✘ %s: %v", j.Lib.Lib.Name, j.Err))
		if j.Output != "" {
			s += "\n" + ui.MutedStyle.MaxWidth(a.width).Render(j.Output)
		}
		return s
	}
	return "" // JobPending: o modal de senha já está na tela
}

func (a *App) renderPassword() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Padding(1, 2)

	content := ui.CursorRowStyle.Render("Senha do sudo") + "\n\n" + a.pwInput.View()
	if a.pwErr != "" {
		content += "\n" + ui.WarnStyle.Render(a.pwErr)
	}
	content += "\n\n" + ui.MutedStyle.Render("enter confirma · esc cancela")
	return box.Render(content)
}

func (a *App) renderSystemInfo() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Padding(1, 2).
		Height(max(a.list.Height()-4, 1))

	content := fmt.Sprintf(
		"Configurações\n\nSistema detectado: %s\nLibs carregadas:   %d",
		a.sys.Type.Label(), len(a.list.Items()),
	)
	return box.Render(content)
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
