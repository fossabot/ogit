package browser

import (
	"context"
	"fmt"
	"log"
	"ogit/internal/gitutils"
	"ogit/internal/utils"
	"ogit/service"
	"path"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func availableKeyBindingsCB() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh list"),
		),
		key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clone a repository (shallow)"),
		),
		key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "browse home page"),
		),
		key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "browse pull requests"),
		),
	}
}

// Update is called whenever the whole model is updated
// It is used for example for messages like "refresh list"
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("Updating UI")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		bottomGap = bottomGap + bottomStatusBarStyle.GetHeight()
		m.list.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)

	case fetchAPIUsageMsg:
		return m, func() tea.Msg {
			apiUsage, err := m.rs.GetAPIUsage(context.Background())
			if err != nil {
				log.Println(err)
				return updateBottomStatusBarMsg(statusError(err.Error()))
			}

			return updateBottomStatusBarMsg(apiUsage.String())
		}
	case updateBottomStatusBarMsg:
		m.bottomStatusBar = string(msg)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "r":
				return m, func() tea.Msg { return refreshReposMsg{} }
			default:
				log.Println("Key Pressed", string(msg.Runes))
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	return m, cmd
}

// delegateItemUpdate is called whenever a specific item is updated.
// It is used for example for messages like "clone repo"
func delegateItemUpdate(cloneDirPath string, orgs []string, rs *service.RepositoryService) list.DefaultDelegate {
	updateFunc := func(msg tea.Msg, m *list.Model) tea.Cmd {
		log.Println("Updating Item")

		selected, ok := m.SelectedItem().(repoListItem)
		if !ok && len(m.VisibleItems()) > 0 {
			return m.NewStatusMessage("unknown item type")
		}

		switch msg := msg.(type) {
		case refreshReposMsg:
			return tea.Batch(
				m.StartSpinner(),
				func() tea.Msg {
					repos, err := rs.GetRepositoriesByOwners(context.Background(), orgs)
					if err != nil {
						log.Println(err)
						return updateStatusMsg(statusError(err.Error()))
					}

					return refreshReposDoneMsg{repos: *repos}
				},
			)

		case refreshReposDoneMsg:
			repos := msg.repos
			newItems := make([]list.Item, len(repos))

			for i := range repos {
				repoItem := repoListItem{
					title:                  repos[i].Owner + "/" + repos[i].Name,
					owner:                  repos[i].Owner,
					name:                   repos[i].Name,
					description:            repos[i].Description,
					browserHomepageURL:     repos[i].BrowserHomepageURL,
					browserPullRequestsURL: repos[i].BrowserPullRequestsURL,
					cloneURL:               repos[i].CloneURL,
				}

				if repoItem.Cloned(cloneDirPath) {
					repoItem.title = statusMessageStyle(repoItem.Title())
					repoItem.description = statusMessageStyle(repoItem.Description())
				}
				newItems[i] = repoItem
			}

			m.SetItems(newItems)
			m.StopSpinner()

			return tea.Batch(
				m.NewStatusMessage(statusMessageStyle(fmt.Sprintf("Fetched %d repos", len(newItems)))),
				func() tea.Msg { return fetchAPIUsageMsg{} },
			)

		case updateStatusMsg:
			m.StopSpinner()
			return m.NewStatusMessage(string(msg))

		case tea.KeyMsg:
			switch msg.String() {
			case "c":
				return tea.Batch(
					m.StartSpinner(),
					func() tea.Msg {
						clonePath := path.Join(cloneDirPath, selected.Owner(), selected.Name())
						if gitutils.Cloned(clonePath) {
							return updateStatusMsg(statusMessageStyle("Already Cloned"))
						}

						repoOnDisk, err := gitutils.CloneToDisk(context.Background(),
							selected.CloneURL(),
							clonePath,
							log.Default().Writer(),
						)
						if err != nil {
							return updateStatusMsg(statusError(err.Error()))
						}

						selected.title = statusMessageStyle(selected.title)
						selected.description = statusMessageStyle(selected.description)

						m.SetItem(m.Index(), selected)
						return updateStatusMsg(statusMessageStyle(repoOnDisk.String()))
					},
				)
			case "w", "p":
				return func() tea.Msg {
					var u string
					if msg.String() == "w" {
						u = selected.BrowserHomepageURL()
					} else if msg.String() == "p" {
						u = selected.BrowserPullRequestsURL()
					}
					if u == "" {
						return updateStatusMsg(statusError("URL not available"))
					}
					err := utils.OpenURL(u)
					if err != nil {
						log.Println(err)
						return updateStatusMsg(statusError(err.Error()))
					}
					return nil
				}

			default:
				lastCommit, err := selected.LastCommitInfo(cloneDirPath)
				if err != nil {
					return m.NewStatusMessage(fmt.Sprintf("unable to read last commit: %s", err))
				}

				return m.NewStatusMessage(lastCommit)
			}
		}

		return nil
	}

	d := list.NewDefaultDelegate()
	d.UpdateFunc = updateFunc
	return d
}
