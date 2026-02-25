package commandpalette

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type resolver struct {
	entries []Entry
	roots   map[string]struct{}
}

type matchResult struct {
	entry Entry
	score int
}

func newResolver(entries []Entry) resolver {
	roots := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if len(entry.Command) == 0 {
			continue
		}
		roots[strings.TrimSpace(entry.Command[0])] = struct{}{}
	}
	return resolver{entries: entries, roots: roots}
}

func (r resolver) Resolve(raw string) (ResolvedCommand, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return ResolvedCommand{}, errors.New("type a command first")
	}

	fields := strings.Fields(candidate)
	if len(fields) == 0 {
		return ResolvedCommand{}, errors.New("no command was detected")
	}
	if strings.HasPrefix(fields[0], "-") {
		return ResolvedCommand{Args: fields, Raw: candidate}, nil
	}
	if _, ok := r.roots[fields[0]]; ok {
		return ResolvedCommand{Args: fields, Raw: candidate}, nil
	}

	query := normalize(candidate)
	matches := make([]matchResult, 0, len(r.entries))
	for _, entry := range r.entries {
		bestScore := 0
		for _, alias := range entry.Aliases {
			score := scoreMatch(query, normalize(alias))
			if score > bestScore {
				bestScore = score
			}
		}
		if bestScore > 0 {
			matches = append(matches, matchResult{entry: entry, score: bestScore})
		}
	}
	if len(matches) == 0 {
		return ResolvedCommand{}, fmt.Errorf("unknown command %q", candidate)
	}

	sort.SliceStable(matches, func(i int, j int) bool {
		if matches[i].score == matches[j].score {
			return len(matches[i].entry.Label) < len(matches[j].entry.Label)
		}
		return matches[i].score > matches[j].score
	})

	selected := matches[0].entry
	return ResolvedCommand{Args: clonePath(selected.Command), Raw: candidate}, nil
}

func scoreMatch(query string, target string) int {
	if query == "" || target == "" {
		return 0
	}
	if query == target {
		return 120
	}
	if strings.HasPrefix(target, query) {
		return 100
	}
	if allTokensPrefix(query, target) {
		return 90
	}
	if strings.Contains(target, query) {
		return 80
	}
	if allTokensContained(query, target) {
		return 70
	}
	return 0
}

func allTokensPrefix(query string, target string) bool {
	queryTokens := strings.Fields(query)
	targetTokens := strings.Fields(target)
	if len(queryTokens) == 0 || len(targetTokens) == 0 {
		return false
	}
	for _, queryToken := range queryTokens {
		matched := false
		for _, targetToken := range targetTokens {
			if strings.HasPrefix(targetToken, queryToken) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func allTokensContained(query string, target string) bool {
	queryTokens := strings.Fields(query)
	targetTokens := strings.Fields(target)
	if len(queryTokens) == 0 || len(targetTokens) == 0 {
		return false
	}
	for _, queryToken := range queryTokens {
		matched := false
		for _, targetToken := range targetTokens {
			if strings.Contains(targetToken, queryToken) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func normalize(value string) string {
	replacer := strings.NewReplacer("_", " ", "-", " ", "/", " ")
	parts := strings.Fields(strings.ToLower(replacer.Replace(strings.TrimSpace(value))))
	return strings.Join(parts, " ")
}
