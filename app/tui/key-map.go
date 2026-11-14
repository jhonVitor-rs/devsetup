package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	confirm    key.Binding
	openConfig key.Binding
	openHelp   key.Binding
	quit       key.Binding

	up     key.Binding
	down   key.Binding
	filter key.Binding
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
		up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir")),
		down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "descer")),
		filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filtrar")),
	}
}

func (k *KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.confirm, k.filter, k.openConfig, k.openHelp, k.quit}
}

func (k *KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.up, k.down, k.filter},
		{k.confirm, k.openConfig, k.openHelp, k.quit},
	}
}
