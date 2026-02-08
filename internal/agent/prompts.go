package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Prompts struct {
	Start      string
	Planner    string
	Critic     string
	CmdAgent   string
	Browser    string
	Researcher string
	Formatter  string
}

func readPromptFile(promptsDir, filename string, required bool) (string, error) {
	if strings.TrimSpace(promptsDir) == "" {
		if required {
			return "", fmt.Errorf("prompts_dir is empty")
		}
		return "", nil
	}
	p := filepath.Join(promptsDir, filename)
	b, err := os.ReadFile(p)
	if err != nil {
		if required {
			return "", fmt.Errorf("failed to read prompt (%s): %w", p, err)
		}
		return "", nil
	}
	return strings.TrimSpace(string(b)), nil
}

func LoadPrompts(promptsDir string) (Prompts, error) {
	start, err := readPromptFile(promptsDir, "start_prompt.txt", true)
	if err != nil {
		return Prompts{}, err
	}
	planner, err := readPromptFile(promptsDir, "planner_prompt.txt", false)
	if err != nil {
		return Prompts{}, err
	}
	critic, err := readPromptFile(promptsDir, "critic_prompt.txt", false)
	if err != nil {
		return Prompts{}, err
	}
	cmdAgent, err := readPromptFile(promptsDir, "cmd_agent_prompt.txt", false)
	if err != nil {
		return Prompts{}, err
	}
	browser, err := readPromptFile(promptsDir, "browser_prompt.txt", false)
	if err != nil {
		return Prompts{}, err
	}
	researcher, err := readPromptFile(promptsDir, "researcher_prompt.txt", false)
	if err != nil {
		return Prompts{}, err
	}
	formatter, err := readPromptFile(promptsDir, "json_formatter.txt", false)
	if err != nil {
		return Prompts{}, err
	}

	return Prompts{
		Start:      start,
		Planner:    planner,
		Critic:     critic,
		CmdAgent:   cmdAgent,
		Browser:    browser,
		Researcher: researcher,
		Formatter:  formatter,
	}, nil
}
