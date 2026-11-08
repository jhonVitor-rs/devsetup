package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	confirm    key.Binding
	openConfig key.Binding
	openHelp   key.Binding
	quit       key.Binding
}

func NewKeyMap() *KeyMap {
	return &KeyMap{
		confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirmar"),
		),
		openConfig: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "abrir configurações"),
		),
		openHelp: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "abrir ajuda"),
		),
		quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancelar/sair"),
		),
	}
}

func (k *KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.confirm, k.openConfig, k.quit,
	}
}

func (k *KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
