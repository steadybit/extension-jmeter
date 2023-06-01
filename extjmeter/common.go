package extjmeter

import (
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"strings"
)

const (
	actionId   = "com.github.steadybit.extension_jmeter.run"
	actionIcon = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjUiIHZpZXdCb3g9IjAgMCAyNCAyNSIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMjAuNzU1IDMuNzQ1Yy0uMTY0LS4xNjMtLjQxLS4yNDUtLjczNy0uMjQ1LS4zMjcuMDgyLTkuMDgyIDEuNTU1LTEyLjY4MiA1LjE1NS0yLjUzNiAyLjUzNi0yLjcgNS4zMTgtMi4yOSA3LjI4MWwzLjAyNy0xLjhjLjQwOS0uMjQ1LjgxOC4zMjguNDkuNjU1bC0zLjAyNyAzLjAyNy0yLjI5IDIuMjkxYS43OS43OSAwIDAwMCAxLjE0NS43NDMuNzQzIDAgMDAuNTcyLjI0NmMuMjQ2IDAgLjQxLS4wODIuNTczLS4yNDVsMi4yOS0yLjI5MWMuNjU1LjMyNyAxLjg4My43MzYgMy40MzcuNzM2IDEuOCAwIDMuNDM3LS41NzMgNC45MS0xLjcxOC4yNDUtLjE2NC40MDgtLjU3My4yNDUtLjlsLTEuMTQ2LTQuMDkxIDIuOTQ2IDEuMTQ1Yy40MDkuMTY0LjkgMCAxLjA2My0uNDA5LjQyNC0uOTQ3LjgwNy0xLjk3NCAxLjE0Ni0yLjk4NkMxOS40NDMgMTAuMjYgMTcuNSA4LjUgMTcuNSA4LjVzMi41LS4wOCAyLjYxNy0uNWMuNTMtMS45MDcuODQtMy4zNTEuODgzLTMuNDM2IDAtLjMyOC0uMDgyLS42NTUtLjI0NS0uODE5eiIgZmlsbD0iY3VycmVudENvbG9yIi8+PC9zdmc+"
)

func stdOutToLog(lines []string) {
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ReplaceAll(line, "\n", ""))
		if len(trimmed) > 0 {
			log.Info().Msgf("---- %s", trimmed)
		}
	}
}

func stdOutToMessages(lines []string) []action_kit_api.Message {
	var messages []action_kit_api.Message
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ReplaceAll(line, "\n", ""))
		if len(trimmed) > 0 {
			messages = append(messages, action_kit_api.Message{
				Level:   extutil.Ptr(action_kit_api.Info),
				Message: trimmed,
			})
		}
	}
	return messages
}
