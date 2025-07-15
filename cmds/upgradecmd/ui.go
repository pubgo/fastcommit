package upgradecmd

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/fastcommit/utils/githubclient"
	"github.com/pubgo/funk/log"
	"github.com/samber/lo"
)

type model struct {
	cursor   int
	assets   []githubclient.Asset
	selected githubclient.Asset
	length   int
}

func initialModel(assets []githubclient.Asset) *model {
	assets = lo.Filter(assets, func(item githubclient.Asset, index int) bool { return !item.IsChecksumFile() })
	sort.Slice(assets, func(i, j int) bool {
		return lo.Must(version.NewSemver(assets[i].Name)).GreaterThan(lo.Must(version.NewSemver(assets[j].Name)))
	})
	return &model{
		assets: assets,
		length: len(assets) - 1,
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp, tea.KeyLeft:
			m.cursor--
			if m.cursor <= 0 {
				m.cursor = 0
			}
		case tea.KeyDown, tea.KeyRight:
			m.cursor++
			if m.cursor == m.length {
				m.cursor = m.length - 1
			}
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

func (m *model) View() string {
	s := "Please Select:\n"

	for i, choice := range m.assets {
		cursor := " "
		if m.cursor%m.length == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %d %s %s %s\n", cursor, i, choice.Name, choice.OS, choice.Arch)
	}

	return s
}
