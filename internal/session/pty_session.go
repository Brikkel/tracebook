package session

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/Brikkel/tracebook/internal/config"
	"github.com/creack/pty"
)

func StartPTYSession(sessionName string, cfg *config.Config) error {
	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}
	mdFile := filepath.Join(cfg.OutputDir, sessionName+".md")
	f, err := os.Create(mdFile)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("# " + sessionName + "\n\n")

	// Determine user's shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	// Start the shell in a PTY
	c := exec.Command(shell, "-l") // -l for login shell
	ptyFile, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer ptyFile.Close()

	// Channels for copying input/output
	go func() {
		io.Copy(os.Stdout, ptyFile) // Shell output to user
	}()
	go func() {
		io.Copy(ptyFile, os.Stdin) // User input to shell
	}()

	// Log commands by tailing the shell's history file
	usr, _ := user.Current()
	histFile := filepath.Join(usr.HomeDir, ".bash_history") // or .zsh_history
	lastSize := int64(0)

	// Build skip command map from config
	skipCmds := make(map[string]bool)
	for _, cmd := range cfg.SkipCommands {
		skipCmds[cmd] = true
	}

	// Main loop: poll for new history entries
	for {
		time.Sleep(1 * time.Second)
		fi, err := os.Stat(histFile)
		if err != nil {
			continue
		}
		if fi.Size() == lastSize {
			continue
		}
		fHist, err := os.Open(histFile)
		if err != nil {
			continue
		}
		fHist.Seek(lastSize, io.SeekStart)
		scanner := bufio.NewScanner(fHist)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			cmdName := strings.Fields(line)[0]
			if skipCmds[cmdName] {
				continue
			}
			// Annotate directory if enabled
			if cfg.AnnotateDirectory {
				cwd, _ := os.Getwd()
				f.WriteString(fmt.Sprintf("`%s`\n", cwd))
			}
			// Note support (if you want to support a note command)
			if strings.HasPrefix(line, "note:") {
				note := strings.TrimSpace(strings.TrimPrefix(line, "note:"))
				f.WriteString(note + "\n\n")
				continue
			}
			f.WriteString("```bash\n" + line + "\n```\n\n")
			// TODO: Add vim diff logic if desired
		}
		lastSize = fi.Size()
		fHist.Close()
		// Check if shell exited
		if c.ProcessState != nil && c.ProcessState.Exited() {
			break
		}
	}
	f.WriteString("\n_Session ended at " + time.Now().Format(time.RFC3339) + "_\n")
	return nil
}
