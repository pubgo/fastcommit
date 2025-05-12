package upgradecmd

import (
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v71/github"
	"github.com/pubgo/funk/log"
	"github.com/samber/lo"
)

type model struct {
	cursor   int
	assets   []*github.ReleaseAsset
	selected *github.ReleaseAsset
	length   int
}

func initialModel(rsp *github.RepositoryRelease) model {
	return model{
		assets: lo.Filter(rsp.Assets, func(item *github.ReleaseAsset, index int) bool {
			return !strings.Contains(item.GetName(), "checksums") && strings.Contains(strings.ToLower(item.GetName()), strings.ToLower(runtime.GOOS))
		}),
		length: len(rsp.Assets) - 1,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp, tea.KeyLeft, tea.KeyDown, tea.KeyRight:
			m.cursor++
		case tea.KeyEnter:
			m.selected = m.assets[m.cursor%m.length]
			return m, tea.Quit
		default:
			log.Error().Str("key", msg.String()).Msg("unknown key")
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Please Select:\n"

	for i, choice := range m.assets {
		cursor := " "
		if m.cursor%m.length == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, lo.FromPtr(choice.Name))
	}

	return s
}
