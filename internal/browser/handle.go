package browser

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/wmalik/ogit/internal/db"
	"github.com/wmalik/ogit/internal/gitconfig"
	"github.com/wmalik/ogit/internal/gitutils"
	"github.com/wmalik/ogit/internal/shell"
	"github.com/wmalik/ogit/internal/sync"

	tea "github.com/charmbracelet/bubbletea"
)

func HandleCommandFetch() error {
	ctx := context.Background()
	gitConf, err := gitconfig.ReadGitConfig()
	if err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll(gitConf.StoragePath(), os.ModePerm); err != nil {
		return err
	}

	localDB, err := db.NewDB(path.Join(gitConf.StoragePath(), "ogit.db"))
	if err != nil {
		log.Fatalln(err)
	}

	if err := localDB.Init(); err != nil {
		log.Fatalln(err)
	}

	if err := sync.Sync(ctx, gitConf); err != nil {
		log.Fatalln(err)
	}

	return nil
}
func HandleCommandDefault() error {
	ctx := context.Background()
	gitConf, err := gitconfig.ReadGitConfig()
	if err != nil {
		log.Fatalln(err)
	}

	localDB, err := db.NewDB(path.Join(gitConf.StoragePath(), "ogit.db"))
	if err != nil {
		log.Fatalln(err)
	}

	if err := localDB.Init(); err != nil {
		log.Fatalln(err)
	}

	repos, err := localDB.SelectAllRepositories(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	gu, err := gitutils.NewGitUtils(gitConf.UseSSHAgent(), gitConf.PrivKeyPath())
	if err != nil {
		log.Fatalln(err)
	}

	f, err := tea.LogToFile(filepath.Join(os.TempDir(), "ogit.log"), "ogit")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	model := NewModelWithItems(repos, gitConf.StoragePath(), gu)
	for {
		if err := tea.NewProgram(model, tea.WithAltScreen()).Start(); err != nil {
			log.Fatalln(err)
		}

		if model.spawnShell {
			model.spawnShell = false
			if err := shell.Spawn(model.selectedItemStoragePath); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}
