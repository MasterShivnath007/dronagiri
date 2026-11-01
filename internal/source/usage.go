package source

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// Last-used detection has to work around two hard facts:
//
//  1. macOS mounts APFS with noatime, so a file's access time never moves.
//  2. Most shells are not recording timestamps in their history.
//
// So we use the best evidence available: Spotlight's own last-used date for
// application bundles, and shell history to prove a command was run at least
// once. Anything we cannot prove is reported as unknown, never guessed.

type usageRecord struct {
	at     time.Time
	seen   bool
	source string
}

var historyStamp = regexp.MustCompile(`^: (\d+):\d+;(.*)$`)

// enrichUsage fills in LastUsed for every tool, in parallel.
func enrichUsage(r *runner.Runner, tools []*model.Tool) {
	history := scanShellHistory()

	var wg sync.WaitGroup
	var mu sync.Mutex
	limit := make(chan struct{}, 8)

	for _, tool := range tools {
		wg.Add(1)
		go func(tool *model.Tool) {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			record := lookupUsage(r, tool, history)

			mu.Lock()
			tool.LastUsed = record.at
			tool.LastUsedSeen = record.seen
			tool.LastUsedSource = record.source
			mu.Unlock()
		}(tool)
	}
	wg.Wait()
}

func lookupUsage(r *runner.Runner, tool *model.Tool, history map[string]usageRecord) usageRecord {
	// Applications: Spotlight knows when macOS last launched them.
	if strings.HasSuffix(tool.Path, ".app") {
		if at, ok := spotlightLastUsed(r, tool.Path); ok {
			return usageRecord{at: at, seen: true, source: "Spotlight"}
		}
	}

	// Commands: look for the package name or any of its binaries.
	candidates := []string{tool.Name}
	for _, binary := range tool.Binaries {
		if base := filepath.Base(binary); base != "" && base != tool.Name {
			candidates = append(candidates, base)
		}
	}

	var best usageRecord
	for _, name := range candidates {
		record, ok := history[name]
		if !ok {
			continue
		}
		if record.at.After(best.at) {
			best.at = record.at
		}
		best.seen = true
		best.source = "shell history"
	}
	return best
}

// scanShellHistory reads the usual history files once and records, for every
// command name it can see, whether it was used and (if the shell stored
// timestamps) when.
func scanShellHistory() map[string]usageRecord {
	out := make(map[string]usageRecord)

	home, err := os.UserHomeDir()
	if err != nil {
		return out
	}

	for _, name := range []string{".zsh_history", ".bash_history", ".sh_history", ".histfile"} {
		readHistoryFile(filepath.Join(home, name), out)
	}

	// zsh keeps a copy of history per session; on many machines those files
	// are the only place older commands survive.
	if sessions, err := filepath.Glob(filepath.Join(home, ".zsh_sessions", "*.history")); err == nil {
		for _, path := range sessions {
			readHistoryFile(path, out)
		}
	}

	return out
}

// readHistoryFile parses one history file in either plain or timestamped
// (`: 1698765432:0;command`) format.
func readHistoryFile(path string, out map[string]usageRecord) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		var at time.Time
		command := line
		if match := historyStamp.FindStringSubmatch(line); match != nil {
			if seconds, err := strconv.ParseInt(match[1], 10, 64); err == nil {
				at = time.Unix(seconds, 0)
			}
			command = match[2]
		}

		for _, commandName := range commandNames(command) {
			record := out[commandName]
			record.seen = true
			if at.After(record.at) {
				record.at = at
			}
			out[commandName] = record
		}
	}
}

// commandNames pulls the first word of each pipeline stage, which is close
// enough to "which programs did you run".
func commandNames(command string) []string {
	segments := strings.FieldsFunc(command, func(r rune) bool {
		return r == ';' || r == '|' || r == '&'
	})

	var names []string
	for _, segment := range segments {
		fields := strings.Fields(segment)
		for _, field := range fields {
			switch {
			case field == "sudo", field == "env", field == "time",
				field == "nohup", field == "command", field == "exec":
				continue
			case strings.HasPrefix(field, "-"):
				continue
			case strings.Contains(field, "="):
				continue // VAR=value prefixes, not the program
			}
			names = append(names, filepath.Base(field))
			break
		}
	}
	return names
}

// spotlightLastUsed asks Spotlight for an application's last launch date.
func spotlightLastUsed(r *runner.Runner, path string) (time.Time, bool) {
	out, err := r.Output("mdls", "-name", "kMDItemLastUsedDate", "-raw", path)
	if err != nil {
		return time.Time{}, false
	}
	out = strings.TrimSpace(out)
	if out == "" || out == "(null)" {
		return time.Time{}, false
	}

	at, err := time.Parse("2006-01-02 15:04:05 -0700", out)
	if err != nil {
		return time.Time{}, false
	}
	return at, true
}
