package libs

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed list/*.toml
var libsFS embed.FS

func Loader() ([]Lib, error) {
	entries, err := fs.ReadDir(libsFS, "list")
	if err != nil {
		return nil, err
	}

	var libs []Lib
	seen := map[string]bool{}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}

		data, err := libsFS.ReadFile(path.Join("list", e.Name()))
		if err != nil {
			return nil, err
		}

		var lib Lib
		if err := toml.Unmarshal(data, &lib); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if lib.ID == "" || lib.Name == "" {
			return nil, fmt.Errorf("%s: id e name são obrigatórios", e.Name())
		}
		if seen[lib.ID] {
			return nil, fmt.Errorf("%s: id duplicado %q", e.Name(), lib.ID)
		}
		seen[lib.ID] = true

		libs = append(libs, lib)
	}

	for _, l := range libs {
		for _, id := range l.Require {
			if id == l.ID {
				return nil, fmt.Errorf("%s: não pode requerer a si mesma", l.ID)
			}
			if !seen[id] {
				return nil, fmt.Errorf("%s: require aponta para %q, que não existe", l.ID, id)
			}
		}
	}
	if err := checkCycles(libs); err != nil {
		return nil, err
	}

	return libs, nil
}

func checkCycles(list []Lib) error {
	deps := map[string][]string{}
	for _, l := range list {
		deps[l.ID] = l.Require
	}

	const (
		visiting = 1
		done     = 2
	)
	state := map[string]int{}

	var visit func(id string) error
	visit = func(id string) error {
		switch state[id] {
		case visiting:
			return fmt.Errorf("dependência circular envolvendo %q", id)
		case done:
			return nil
		}
		state[id] = visiting
		for _, dep := range deps[id] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		state[id] = done
		return nil
	}

	for _, l := range list {
		if err := visit(l.ID); err != nil {
			return err
		}
	}
	return nil
}
