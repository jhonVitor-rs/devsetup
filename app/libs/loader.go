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
	return libs, nil
}
