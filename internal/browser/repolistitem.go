package browser

import (
	"ogit/internal/gitutils"
	"path"
)

type repoListItem struct {
	title                  string
	owner                  string
	name                   string
	description            string
	browserHomepageURL     string
	browserPullRequestsURL string
	httpsCloneURL          string
	sshCloneURL            string
}

func (i repoListItem) Title() string                  { return i.title }
func (i repoListItem) Owner() string                  { return i.owner }
func (i repoListItem) Name() string                   { return i.name }
func (i repoListItem) Description() string            { return i.description }
func (i repoListItem) FilterValue() string            { return i.title + i.description }
func (i repoListItem) BrowserHomepageURL() string     { return i.browserHomepageURL }
func (i repoListItem) BrowserPullRequestsURL() string { return i.browserPullRequestsURL }
func (i repoListItem) HTTPSCloneURL() string          { return i.httpsCloneURL }
func (i repoListItem) SSHCloneURL() string            { return i.sshCloneURL }
func (i repoListItem) Cloned(cloneDirPath string) bool {
	return gitutils.Cloned(path.Join(cloneDirPath, i.owner, i.name))
}

func (i repoListItem) LastCommitInfo(cloneDirPath string) (string, error) {
	if i.Cloned(cloneDirPath) {
		repo, err := gitutils.ReadRepository(path.Join(cloneDirPath, i.owner, i.name))
		if err != nil {
			return "", err
		}

		return repo.LastCommit(), nil
	}

	return "", nil
}
