package system

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

type SystemType string

const (
	Windows  SystemType = "windows"
	Mac      SystemType = "mac"
	LinuxDeb SystemType = "deb"
	LinuxArc SystemType = "arc"
)

type System struct {
	Type SystemType
}

func New() (*System, error) {
	t, err := detectSystem()
	if err != nil {
		return nil, err
	}

	return &System{Type: t}, nil
}

func TestParseOSRelease(t *testing.T) {
	cases := map[string]struct {
		in   string
		want SystemType
	}{
		"manjaro": {"ID=manjaro\nID_LIKE=arch\n", LinuxArc},
		"ubuntu":  {"ID=ubuntu\nID_LIKE=debian\n", LinuxDeb},
		"mint":    {"ID=linuxmint\nID_LIKE=\"ubuntu debian\"\n", LinuxDeb},
		"arch":    {"ID=arch\n", LinuxArc},
	}
	for name, c := range cases {
		got, err := parseOSRelease(strings.NewReader(c.in))
		if err != nil || got != c.want {
			t.Errorf("%s: got %v, %v; want %v", name, got, err, c.want)
		}
	}
}

func detectSystem() (SystemType, error) {
	switch runtime.GOOS {
	case "windows":
		return Windows, nil
	case "darwin":
		return Mac, nil
	case "linux":
		return detectLinux()
	}
	return "", fmt.Errorf("sistema não suportado: %s", runtime.GOOS)
}

func detectLinux() (SystemType, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return "", err
	}
	defer f.Close()
	return parseOSRelease(f)
}

func parseOSRelease(r io.Reader) (SystemType, error) {
	var id, idLike string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if !ok {
			continue
		}
		v = strings.Trim(v, `"'`)
		switch k {
		case "ID":
			id = v
		case "ID_LIKE":
			idLike = v
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}

	all := id + " " + idLike
	switch {
	case strings.Contains(all, "arch"):
		return LinuxArc, nil
	case strings.Contains(all, "debian"), strings.Contains(all, "ubuntu"):
		return LinuxDeb, nil
	}
	return "", fmt.Errorf("distro não suportada: %s", id)
}
