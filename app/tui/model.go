package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jhonVitor-rs/devsetup/app/libs"
	"github.com/jhonVitor-rs/devsetup/app/system"
)

type JobState int

const (
	JobPending JobState = iota
	JobRunning
	JobDone
	JobFailed
)

type Job struct {
	Lib   libs.LibStatus
	State JobState
	Err   error
}

type InstallRequestedMsg struct{ Lib libs.LibStatus }
type installDoneMsg struct {
	ID  string
	Err error
}
type recheckedMsg struct{ Status libs.LibStatus }

type App struct {
	sys *system.System

	list    list.Model
	spinner spinner.Model
	help    help.Model
	keys    *KeyMap

	queue          []*Job
	showHelp       bool
	showSystemInfo bool
	width, height  int
}

func New(sys *system.System, statuses []libs.LibStatus) *App {
	keys := NewKeyMap()
	a := &App{
		sys:     sys,
		keys:    keys,
		help:    help.New(),
		spinner: spinner.New(),
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.list.SetSize(msg.Width, msg.Height-4)

	case InstallRequestedMsg:
		id := msg.Lib.Lib.ID
		if a.active(id) {
			return a, nil
		}
		a.dropFinished(id)
		a.queue = append(a.queue, &Job{Lib: msg.Lib, State: JobPending, Err: nil})
		cmds = append(cmds, a.startNext())

	case installDoneMsg:
		job := a.job(msg.ID)
		if job == nil {
			return a, nil
		}
		job.Err = msg.Err
		job.State = JobDone
		if msg.Err != nil {
			job.State = JobFailed
		}
		cmds = append(cmds, a.recheck(job.Lib.Lib), a.startNext())

	case recheckedMsg:
		cmds = append(cmds, a.updateStatus(msg.Status))
	}

	// repassa a mensagem aos componentes
	var cmd tea.Cmd
	a.list, cmd = a.list.Update(msg)
	cmds = append(cmds, cmd)

	a.spinner, cmd = a.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...) // Batch ignora cmds nil
}

func (a *App) View() string {
	out := a.list.View()
	for _, j := range a.queue {
		if j.State == JobFailed {
			out += fmt.Sprintf("\n✘ %s: %v", j.Lib.Lib.Name, j.Err)
		}
	}
	return out
}

func (a *App) jobState(id string) (JobState, bool) {
	if j := a.job(id); j != nil {
		return j.State, true
	}
	return 0, false
}

func (a *App) active(id string) bool {
	for _, j := range a.queue {
		if j.Lib.Lib.ID == id && (j.State == JobPending || j.State == JobRunning) {
			return true
		}
	}
	return false
}

func (a *App) dropFinished(id string) {
	kept := a.queue[:0]
	for _, j := range a.queue {
		if j.Lib.Lib.ID == id && (j.State == JobDone || j.State == JobFailed) {
			continue
		}
		kept = append(kept, j)
	}
	a.queue = kept
}

func (a *App) job(id string) *Job {
	for _, j := range a.queue {
		if j.Lib.Lib.ID == id {
			return j
		}
	}
	return nil
}

func (a *App) startNext() tea.Cmd {
	for _, j := range a.queue {
		if j.State == JobRunning {
			return nil
		}
	}
	for _, j := range a.queue {
		if j.State != JobPending {
			continue
		}
		cmd, err := a.sys.InstallCmd(j.Lib.Lib)
		if err != nil {
			j.State, j.Err = JobFailed, err
			continue
		}
		j.State = JobRunning
		id := j.Lib.Lib.ID
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return installDoneMsg{ID: id, Err: err}
		})
	}
	return nil
}

func (a *App) recheck(lib libs.Lib) tea.Cmd {
	sys := a.sys
	return func() tea.Msg { return recheckedMsg{Status: sys.RunCheck(lib)} }
}

func (a *App) updateStatus(st libs.LibStatus) tea.Cmd {
	for i, it := range a.list.Items() {
		if ls, ok := it.(libs.LibStatus); ok && ls.Lib.ID == st.Lib.ID {
			return a.list.SetItem(i, st)
		}
	}
	return nil
}
