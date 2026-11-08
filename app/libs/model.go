package libs

type Lib struct {
	ID      string            `toml:"id"`
	Name    string            `toml:"name"`
	Check   string            `toml:"check"`
	Install map[string]string `toml:"install"`
}

type LibStatus struct {
	Lib       Lib
	Installed bool
	Version   string
}

func (l LibStatus) ToggleInstalled() LibStatus {
	l.Installed = !l.Installed
	return l
}

func (l LibStatus) FilterValue() string { return l.Lib.ID }
