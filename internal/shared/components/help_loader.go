package components

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"ldt-toolkit-cli/internal/shared/model"
)

const defaultHelpUvBinary = "uv"

type CustomHelpNodeLoader func(path []string) (*model.ParsedHelp, bool, error)

type HelpLoaderConfig struct {
	UvBinary          string
	EntryPoint        string
	Environment       []string
	ProjectRootFinder func() (string, error)
	CustomNodeLoader  CustomHelpNodeLoader
}

type HelpLoader struct {
	mu     sync.Mutex
	cache  map[string]*model.ParsedHelp
	config HelpLoaderConfig
}

var (
	helpANSIRegex      = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	helpCommandPattern = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9_-]*)\s{2,}(.+)$`)
)

func NewHelpLoader(config HelpLoaderConfig) *HelpLoader {
	cfg := config
	cfg.UvBinary = strings.TrimSpace(cfg.UvBinary)
	if cfg.UvBinary == "" {
		cfg.UvBinary = defaultHelpUvBinary
	}
	cfg.EntryPoint = strings.TrimSpace(cfg.EntryPoint)
	if cfg.EntryPoint == "" {
		cfg.EntryPoint = "ldt-toolkit"
	}
	if len(cfg.Environment) == 0 {
		cfg.Environment = os.Environ()
	}

	return &HelpLoader{
		cache:  make(map[string]*model.ParsedHelp),
		config: cfg,
	}
}

func (l *HelpLoader) Load(path []string) (*model.ParsedHelp, error) {
	key := model.JoinPath(path)

	if l.config.CustomNodeLoader != nil {
		if customNode, handled, err := l.config.CustomNodeLoader(path); handled {
			if err != nil {
				return nil, err
			}
			l.mu.Lock()
			l.cache[key] = customNode
			l.mu.Unlock()
			return customNode, nil
		}
	}

	l.mu.Lock()
	cached, ok := l.cache[key]
	l.mu.Unlock()
	if ok {
		return cached, nil
	}

	args := append(model.ClonePath(path), "--help")
	commandArgs := append([]string{"run", l.config.EntryPoint}, args...)
	command := exec.Command(l.config.UvBinary, commandArgs...)
	command.Env = l.config.Environment
	if l.config.ProjectRootFinder != nil {
		if rootDir, rootErr := l.config.ProjectRootFinder(); rootErr == nil {
			command.Dir = rootDir
		}
	}

	output, err := command.CombinedOutput()
	text := normaliseHelpText(string(output))
	if err != nil {
		return nil, fmt.Errorf("subcommand discovery failed for `%s`: %w\n%s", strings.Join(path, " "), err, text)
	}

	parsed := parseHelpOutput(text)
	parsed.Path = model.ClonePath(path)
	parsed.Raw = text

	l.mu.Lock()
	l.cache[key] = &parsed
	l.mu.Unlock()

	return &parsed, nil
}

func parseHelpOutput(text string) model.ParsedHelp {
	lines := strings.Split(text, "\n")

	usage := ""
	summary := ""
	commands := make([]model.CommandDef, 0)

	usageIndex := -1
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Usage:") {
			usageIndex = index
			usageParts := []string{strings.TrimSpace(strings.TrimPrefix(trimmed, "Usage:"))}
			for next := index + 1; next < len(lines); next++ {
				candidate := strings.TrimSpace(lines[next])
				if candidate == "" {
					break
				}
				if strings.HasPrefix(candidate, "╭") || strings.HasPrefix(candidate, "│") || strings.HasPrefix(candidate, "╰") {
					break
				}
				usageParts = append(usageParts, candidate)
			}
			usage = strings.Join(usageParts, " ")
			break
		}
	}

	if usageIndex >= 0 {
		summaryParts := make([]string, 0)
		for index := usageIndex + 1; index < len(lines); index++ {
			candidate := strings.TrimSpace(lines[index])
			if candidate == "" {
				if len(summaryParts) > 0 {
					break
				}
				continue
			}
			if strings.HasPrefix(candidate, "╭") || strings.HasPrefix(candidate, "│") || strings.HasPrefix(candidate, "╰") {
				if len(summaryParts) > 0 {
					break
				}
				continue
			}
			if strings.HasPrefix(candidate, "Try '") {
				continue
			}
			summaryParts = append(summaryParts, candidate)
		}
		summary = strings.Join(summaryParts, " ")
	}

	inSection := false
	currentIndex := -1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "╭─") {
			inSection = true
			currentIndex = -1
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "╰") {
			inSection = false
			currentIndex = -1
			continue
		}
		if !inSection {
			continue
		}

		left := strings.Index(line, "│")
		right := strings.LastIndex(line, "│")
		if left < 0 || right <= left {
			continue
		}

		rawCell := strings.TrimRight(line[left+len("│"):right], " ")
		if strings.TrimSpace(rawCell) == "" {
			continue
		}

		name, desc, ok := parseCommandRow(rawCell)
		if ok {
			commands = append(commands, model.CommandDef{
				Name:        name,
				DisplayName: name,
				Description: desc,
			})
			currentIndex = len(commands) - 1
			continue
		}

		if currentIndex >= 0 {
			continuation := strings.TrimSpace(rawCell)
			if continuation != "" {
				commands[currentIndex].Description = strings.TrimSpace(commands[currentIndex].Description + " " + continuation)
			}
		}
	}

	return model.ParsedHelp{
		Usage:    usage,
		Summary:  summary,
		Commands: dedupeCommands(commands),
	}
}

func parseCommandRow(content string) (string, string, bool) {
	trimmed := strings.TrimLeft(content, " ")
	matches := helpCommandPattern.FindStringSubmatch(trimmed)
	if len(matches) != 3 {
		return "", "", false
	}
	name := strings.TrimSpace(matches[1])
	description := strings.TrimSpace(matches[2])
	if name == "" || description == "" {
		return "", "", false
	}
	return name, description, true
}

func dedupeCommands(commands []model.CommandDef) []model.CommandDef {
	seen := make(map[string]bool, len(commands))
	filtered := make([]model.CommandDef, 0, len(commands))
	for _, command := range commands {
		if seen[command.Name] {
			continue
		}
		seen[command.Name] = true
		filtered = append(filtered, command)
	}
	return filtered
}

func normaliseHelpText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = helpANSIRegex.ReplaceAllString(value, "")
	return strings.TrimSpace(value)
}
