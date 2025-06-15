package session

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Brikkel/tracebook/internal/config"
	"github.com/Brikkel/tracebook/internal/diff"
)

var skipCmds = map[string]bool{
	"cd": true, "ls": true, "pwd": true, "clear": true,
}

func Start(sessionName string, cfg *config.Config) {
	mdFile := sessionName + ".md"
	f, err := os.Create(mdFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("# " + sessionName + "\n\n")

	reader := bufio.NewReader(os.Stdin)
	var lastVimFile string
	var lastVimContent []byte

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

		if strings.HasPrefix(line, "note:") {
			note := strings.TrimSpace(strings.TrimPrefix(line, "note:"))
			f.WriteString(note + "\n\n")
			continue
		}

		cmdName := strings.Fields(line)[0]
		if skipCmds[cmdName] {
			continue
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
