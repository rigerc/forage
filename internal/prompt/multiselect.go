package prompt

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type multiItem struct {
	name string
}

func (i multiItem) FilterValue() string { return i.name }

type multiDelegate struct {
	selected          map[int]bool
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
	checkedStyle      lipgloss.Style
	uncheckedStyle    lipgloss.Style
}

func (d multiDelegate) Height() int                             { return 1 }
func (d multiDelegate) Spacing() int                            { return 0 }
func (d multiDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d multiDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(multiItem)
	if !ok {
		return
	}

	checkbox := "[ ]"
	if d.selected[index] {
		checkbox = d.checkedStyle.Render("[x]")
	} else {
		checkbox = d.uncheckedStyle.Render("[ ]")
	}

	text := checkbox + " " + i.name

	fn := d.itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.selectedItemStyle.Render(strings.Join(s, ""))
		}
	}

	fmt.Fprint(w, fn(text))
}

type multiSelectModel struct {
	list     list.Model
	selected map[int]bool
	items    []multiItem
	quitting bool
	header   string
}

func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case " ":
			idx := m.list.Index()
			if m.selected[idx] {
				delete(m.selected, idx)
			} else {
				m.selected[idx] = true
			}
			return m, nil
		case "enter":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m multiSelectModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder
	b.WriteString(m.header)
	b.WriteString("\n")
	b.WriteString(m.list.View())
	b.WriteString("\n  ")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("space: toggle  enter: confirm  esc: cancel"))
	b.WriteString("\n")

	return tea.NewView(b.String())
}

func RunMultiSelect(header string, items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}

	multiItems := make([]multiItem, len(items))
	listItems := make([]list.Item, len(items))
	for i, name := range items {
		mi := multiItem{name: name}
		multiItems[i] = mi
		listItems[i] = mi
	}

	delegate := multiDelegate{
		selected:          make(map[int]bool),
		itemStyle:         lipgloss.NewStyle(),
		selectedItemStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true),
		checkedStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		uncheckedStyle:    lipgloss.NewStyle().Faint(true),
	}

	height := len(items)*2 + 4
	if height < 6 {
		height = 6
	}

	l := list.New(listItems, delegate, 60, height)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.KeyMap.Quit.SetEnabled(false)
	l.KeyMap.Filter.SetEnabled(false)

	delegate.selected = make(map[int]bool)

	m := multiSelectModel{
		list:     l,
		selected: delegate.selected,
		items:    multiItems,
		header:   boldStyle.Render(header),
	}

	p := tea.NewProgram(m)
	tm, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running multi-select: %w", err)
	}

	result := tm.(multiSelectModel)
	if result.quitting {
		return nil, fmt.Errorf("cancelled")
	}

	var selected []string
	for idx := range result.selected {
		if idx >= 0 && idx < len(result.items) {
			selected = append(selected, result.items[idx].name)
		}
	}

	return selected, nil
}

var boldStyle = lipgloss.NewStyle().Bold(true)
