package components

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultHintPauseEvery = 1 * time.Second
	defaultHintTypeEvery  = 60 * time.Millisecond
)

type rotatingHintPauseMsg struct {
	cycleID uint64
}

type rotatingHintTypeMsg struct {
	cycleID uint64
}

// RotatingHint cycles through example commands and exposes a typewriter-style
// display string that is progressively revealed character by character.
type RotatingHint struct {
	values []string

	currentIndex int

	typedChars int
	typing     bool

	pauseEvery time.Duration
	typeEvery  time.Duration

	rng *rand.Rand

	cycleID uint64
}

func NewRotatingHint(values []string) RotatingHint {
	cleaned := cleanRotatingHintValues(values)

	seed := time.Now().UnixNano()
	if seed == 0 {
		seed = 1
	}

	hint := RotatingHint{
		values:       cleaned,
		currentIndex: 0,
		typedChars:   0,
		typing:       false,
		pauseEvery:   defaultHintPauseEvery,
		typeEvery:    defaultHintTypeEvery,
		rng:          rand.New(rand.NewSource(seed)),
		cycleID:      0,
	}

	if len(cleaned) > 1 {
		start := hint.rng.Intn(len(cleaned))
		hint.currentIndex = start
	}
	hint.typedChars = hint.initialTypedChars(hint.Current())

	return hint
}

func (h *RotatingHint) Reset() tea.Cmd {
	h.cycleID++

	if len(h.values) > 1 {
		start := h.rng.Intn(len(h.values))
		h.currentIndex = start
	} else {
		h.currentIndex = 0
	}
	h.typedChars = h.initialTypedChars(h.Current())
	h.typing = h.typedChars < h.currentRuneCount()

	if h.typing {
		return h.scheduleType()
	}
	return h.schedulePause()
}

// Update returns whether the displayed text changed and the follow-up command.
func (h *RotatingHint) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch typed := msg.(type) {
	case rotatingHintPauseMsg:
		if typed.cycleID != h.cycleID {
			return false, nil
		}
		if len(h.values) < 2 {
			return false, nil
		}

		h.currentIndex = h.nextIndex()
		h.typedChars = h.initialTypedChars(h.Current())
		h.typing = h.typedChars < h.currentRuneCount()

		if h.typing {
			return true, h.scheduleType()
		}
		return true, h.schedulePause()

	case rotatingHintTypeMsg:
		if typed.cycleID != h.cycleID {
			return false, nil
		}
		if !h.typing {
			return false, nil
		}

		total := h.currentRuneCount()
		if total <= 0 {
			h.typing = false
			return false, h.schedulePause()
		}

		if h.typedChars < total {
			h.typedChars++
		}

		if h.typedChars >= total {
			h.typing = false
			return true, h.schedulePause()
		}
		return true, h.scheduleType()
	}

	return false, nil
}

func (h RotatingHint) Current() string {
	if len(h.values) == 0 {
		return "data conversion"
	}
	if h.currentIndex < 0 || h.currentIndex >= len(h.values) {
		return h.values[0]
	}
	return h.values[h.currentIndex]
}

// Display returns the currently typed portion of the active command.
func (h RotatingHint) Display() string {
	command := h.Current()
	if command == "" {
		return ""
	}
	runes := []rune(command)
	if h.typedChars <= 0 {
		return ""
	}
	if h.typedChars >= len(runes) {
		return string(runes)
	}
	return string(runes[:h.typedChars])
}

func (h RotatingHint) nextIndex() int {
	if len(h.values) <= 1 {
		return 0
	}

	next := h.rng.Intn(len(h.values) - 1)
	if next >= h.currentIndex {
		next++
	}
	return next
}

func (h RotatingHint) schedulePause() tea.Cmd {
	cycleID := h.cycleID
	return tea.Tick(h.pauseEvery, func(time.Time) tea.Msg {
		return rotatingHintPauseMsg{cycleID: cycleID}
	})
}

func (h RotatingHint) scheduleType() tea.Cmd {
	cycleID := h.cycleID
	return tea.Tick(h.typeEvery, func(time.Time) tea.Msg {
		return rotatingHintTypeMsg{cycleID: cycleID}
	})
}

func cleanRotatingHintValues(values []string) []string {
	seen := map[string]struct{}{}
	cleaned := make([]string, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		cleaned = append(cleaned, trimmed)
	}

	if len(cleaned) == 0 {
		return []string{"data conversion"}
	}
	return cleaned
}

func (h RotatingHint) currentRuneCount() int {
	return len([]rune(h.Current()))
}

func (h RotatingHint) initialTypedChars(command string) int {
	if strings.TrimSpace(command) == "" {
		return 0
	}
	if len([]rune(command)) == 0 {
		return 0
	}
	return 1
}
