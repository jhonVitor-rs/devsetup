package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jhonVitor-rs/devsetup/app/libs"
)

func (s *System) ShellCmd(ctx context.Context, command string) *exec.Cmd {
	if s.Type == Windows {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func (s *System) NeedsSudo(lib libs.Lib) bool {
	return s.Type != Windows && strings.Contains(lib.Install[string(s.Type)], "sudo")
}

func (s *System) CheckSudo(password string) bool {
	cmd := exec.Command("sudo", "-S", "-p", "", "-v")
	cmd.Stdin = strings.NewReader(password + "\n")
	return cmd.Run() == nil
}

func (s *System) RunCheck(lib libs.Lib) libs.LibStatus {
	st := libs.LibStatus{Lib: lib}
	if lib.Check == "" {
		return st
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := s.ShellCmd(ctx, lib.Check).CombinedOutput()
	if err == nil {
		st.Installed = true
		st.Version = strings.TrimSpace(string(out))
	}

	return st
}

func (s *System) CheckAll(list []libs.Lib) []libs.LibStatus {
	out := make([]libs.LibStatus, len(list))
	var wg sync.WaitGroup
	for i, l := range list {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = s.RunCheck(l)
		}()
	}
	wg.Wait()
	return out
}

func (s *System) RunInstall(lib libs.Lib, password string) (string, error) {
	command, ok := lib.Install[string(s.Type)]
	if !ok || command == "" {
		return "", fmt.Errorf("%s não tem instalação definida para %s", lib.Name, s.Type)
	}

	script := command
	if s.NeedsSudo(lib) {
		script = `sudo() { command sudo -S -p '' "$@"; }; ` + command
	}

	cmd := s.ShellCmd(context.Background(), script)
	if password != "" {
		cmd.Stdin = strings.NewReader(password + "\n")
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}
