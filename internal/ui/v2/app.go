package v2

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nkzk/xrefs/internal/ui/v2/models"
	"github.com/nkzk/xrefs/internal/ui/v2/store"
)

type Config struct {
	ResourceName    string
	ResourceGroup   string
	ResourceVersion string
	Name            string
	Namespace       string
	Title           string
}

type App struct {
	config Config
	store  *store.Store
	model  *models.RootModel
}

func New(cfg Config) *App {
	storeCfg := store.Config{
		ResourceName:    cfg.ResourceName,
		ResourceGroup:   cfg.ResourceGroup,
		ResourceVersion: cfg.ResourceVersion,
		Name:            cfg.Name,
		Namespace:       cfg.Namespace,
	}
	s := store.NewStore(storeCfg, store.NewKubectlClient())
	root := models.NewRootModel(s, cfg.Title)

	return &App{
		config: cfg,
		store:  s,
		model:  root,
	}
}

func NewWithMockClient(cfg Config, client store.Client) *App {
	storeCfg := store.Config{
		ResourceName:    cfg.ResourceName,
		ResourceGroup:   cfg.ResourceGroup,
		ResourceVersion: cfg.ResourceVersion,
		Name:            cfg.Name,
		Namespace:       cfg.Namespace,
	}
	s := store.NewStore(storeCfg, client)
	root := models.NewRootModel(s, cfg.Title)

	return &App{
		config: cfg,
		store:  s,
		model:  root,
	}
}

func (a *App) Run() error {
	p := tea.NewProgram(a.model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) Model() tea.Model {
	return a.model
}
