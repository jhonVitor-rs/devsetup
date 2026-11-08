package tui

import (
	"fmt"
	"io"
	"regexp"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jhonVitor-rs/devsetup/app/libs"
	"github.com/jhonVitor-rs/devsetup/app/ui"
)

var (
	versionRe = regexp.MustCompile(`\d+(\.\d+)+`)
)

type Delegate struct {
	keys     *KeyMap
	jobState func(id string) (JobState, bool)
}

func NewDelegate(keys *KeyMap, jobState func(id string) (JobState, bool)) *Delegate {
	return &Delegate{
		keys:     keys,
		jobState: jobState,
	}
}

func (d *Delegate) Height() int  { return 2 }
func (d *Delegate) Spacing() int { return 1 }

func (d *Delegate) busy(id string) bool {
	st, ok := d.jobState(id)
	return ok && (st == JobPending || st == JobRunning)
}

func (d *Delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || m.FilterState() == list.Filtering {
		return nil
	}
	if !key.Matches(keyMsg, d.keys.confirm) {
		return nil
	}

	lib, ok := m.SelectedItem().(libs.LibStatus)
	if !ok || lib.Installed || d.busy(lib.Lib.ID) {
		return nil
	}

	return func() tea.Msg { return InstallRequestedMsg{Lib: lib} }
}

func (d *Delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	lib, ok := item.(libs.LibStatus)
	if !ok {
		return
	}

	marker, name := "  ", lib.Lib.Name
	if index == m.Index() {
		marker, name = ui.CursorRowStyle.Render("▌ "), ui.CursorRowStyle.Render(name)
	}

	icon, desc := ui.MutedStyle.Render("○"), "não instalada"
	switch {
	case lib.Installed:
		icon, desc = ui.InstalledStyle.Render("✔"), "instalada"
		if v := versionRe.FindString(lib.Version); v != "" {
			desc = "v" + v
		}
	case d.busy(lib.Lib.ID):
		icon, desc = ui.WarnStyle.Render("…"), "instalando…"
	}

	fmt.Fprintf(w, "%s%s %s\n    %s", marker, icon, name, ui.MutedStyle.Render(desc))
}
