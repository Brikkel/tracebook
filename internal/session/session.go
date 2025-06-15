package session

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Brikkel/tracebook/internal/config"
	"github.com/Brikkel/tracebook/internal/diff"
)

func Start(sessionName string, cfg *config.Config) {
	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		panic(err)
	}
	mdFile := filepath.Join(cfg.OutputDir, sessionName+".md")
	f, err := os.Create(mdFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("# " + sessionName + "\n\n")

	reader := bufio.NewReader(os.Stdin)
	var lastVimFile string
	var lastVimContent []byte

	// Build skip command map from config
	skipCmds := make(map[string]bool)
	for _, cmd := range cfg.SkipCommands {
		skipCmds[cmd] = true
	}

	for {
		fmt.Print("$ ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" || line == ":q" {
			fmt.Println("Session ended.")
			f.WriteString("\n_Session ended at " + time.Now().Format(time.RFC3339) + "_\n")
			break
		}

		if strings.HasPrefix(line, "note:") {
			note := strings.TrimSpace(strings.TrimPrefix(line, "note:"))
			f.WriteString(note + "\n\n")
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

		if cmdName == "vim" && len(strings.Fields(line)) > 1 {
			vimFile := strings.Fields(line)[1]
			before, _ := os.ReadFile(vimFile)
			lastVimFile = vimFile
			lastVimContent = before
		}

		f.WriteString("```bash\n" + line + "\n```\n\n")

		if cmdName == "vim" && len(strings.Fields(line)) > 1 {
			cmd := exec.Command("vim", strings.Fields(line)[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()

			after, _ := os.ReadFile(lastVimFile)
			if string(lastVimContent) != string(after) {
				diffText := diff.GetDiff(string(lastVimContent), string(after))
				f.WriteString("```diff\n" + diffText + "\n```\n\n")
			}
			lastVimFile = ""
			lastVimContent = nil
		} else {
			cmd := exec.Command("bash", "-c", line)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
}
