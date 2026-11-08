package app

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jhonVitor-rs/devsetup/app/libs"
	"github.com/jhonVitor-rs/devsetup/app/system"
	"github.com/jhonVitor-rs/devsetup/app/tui"
)

// load é compartilhado pela TUI e pelos comandos não interativos.
func load() (*system.System, []libs.LibStatus, error) {
	sys, err := system.New()
	if err != nil {
		return nil, nil, fmt.Errorf("detectar sistema: %w", err)
	}
	defs, err := libs.Loader()
	if err != nil {
		return nil, nil, fmt.Errorf("carregar libs: %w", err)
	}
	return sys, sys.CheckAll(defs), nil
}

func Run() error {
	sys, statuses, err := load()
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(tui.New(sys, statuses), tea.WithAltScreen()).Run()
	return err
}

func List(w io.Writer) error {
	_, statuses, err := load()
	if err != nil {
		return err
	}
	for _, s := range statuses {
		mark := "○"
		if s.Installed {
			mark = "✔"
		}
		fmt.Fprintf(w, "%s %s\n", mark, s.Lib.Name)
	}
	return nil
}
