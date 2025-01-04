package fastcommit

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	semver "github.com/hashicorp/go-version"
)

type model2 struct {
	textInput textinput.Model
	exit      bool
}

// sanitizeInput verifies that an input text string gets validated
func sanitizeInput(input string) error {
	_, err := semver.NewSemver(input)
	return err
}

func InitialTextInputModel(data string) model2 {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = ""
	ti.CharLimit = len(data) + 20
	ti.Width = len(data) + 20
	ti.Validate = sanitizeInput
	ti.SetValue(data)

	return model2{
		textInput: ti,
	}
}

// Init is called at the beginning of a textinput step
// and sets the cursor to blink
func (m model2) Init() tea.Cmd {
	return textinput.Blink
}

// Update is called when "things happen", it checks for the users text input,
// and for Ctrl+C or Esc to close the program.
func (m model2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.exit = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View is called to draw the textinput step
func (m model2) View() string {
	return fmt.Sprintf(
		"git message: %s\n",
		m.textInput.View(),
	)
}

func (m model2) Value() string {
	return m.textInput.Value()
}
func (m model2) isExit() bool {
	return m.exit
}
