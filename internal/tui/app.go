package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/tui/components"
)

// PageID identifies a TUI page.
type PageID int

const (
	PageHome PageID = iota
	PageServices
	PageServiceDetail
	PageProviders
	PageSecurity
	PageFeatures
	PageVersion
	PageEditor
	PageImport
	PageDeploy
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// page is the interface each TUI page must implement.
type page interface {
	Update(msg tea.Msg) (page, tea.Cmd)
	View() string
}

// App is the root bubbletea model and page router.
type App struct {
	cfg         *config.SpinctlConfig
	configPath  string
	version     string
	currentPage PageID
	pageStack   []PageID
	homePage    page
	servicesPage page
	editorPage  page
	importPage  page
	deployPage  page
	dirty       bool
	width       int
	height      int
	statusBar   *components.StatusBar
}

// NewApp creates a new App with the home page active.
func NewApp(cfg *config.SpinctlConfig, configPath string, version string) *App {
	app := &App{
		cfg:         cfg,
		configPath:  configPath,
		version:     version,
		currentPage: PageHome,
		width:       80,
		height:      24,
		statusBar:   components.NewStatusBar(80),
	}
	app.homePage = NewHomePage(cfg)
	return app
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar = components.NewStatusBar(a.width)
		a.statusBar.SetModified(a.dirty)
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "esc":
			if a.currentPage != PageHome {
				a.goBack()
				return a, nil
			}
		}
	}

	// Delegate to current page.
	var cmd tea.Cmd
	switch a.currentPage {
	case PageHome:
		if a.homePage != nil {
			var updated page
			updated, cmd = a.homePage.Update(msg)
			a.homePage = updated

			// Check if home page selected something.
			if hp, ok := a.homePage.(*HomePage); ok && hp.selected != 0 {
				a.navigateTo(hp.selected)
				hp.selected = 0
			}
		}
	case PageServices:
		if a.servicesPage != nil {
			var updated page
			updated, cmd = a.servicesPage.Update(msg)
			a.servicesPage = updated
		}
	case PageProviders, PageSecurity, PageFeatures, PageVersion, PageEditor:
		if a.editorPage != nil {
			var updated page
			updated, cmd = a.editorPage.Update(msg)
			a.editorPage = updated
		}
	case PageImport:
		if a.importPage != nil {
			var updated page
			updated, cmd = a.importPage.Update(msg)
			a.importPage = updated
		}
	case PageDeploy:
		if a.deployPage != nil {
			var updated page
			updated, cmd = a.deployPage.Update(msg)
			a.deployPage = updated
		}
	}

	return a, cmd
}

func (a *App) navigateTo(pageID PageID) {
	a.pageStack = append(a.pageStack, a.currentPage)
	a.currentPage = pageID

	switch pageID {
	case PageServices:
		a.servicesPage = NewServicesPage(a.cfg)
	case PageProviders:
		a.editorPage = NewProvidersPage(a.cfg)
	case PageSecurity:
		a.editorPage = NewSecurityPage(a.cfg)
	case PageFeatures:
		a.editorPage = NewFeaturesPage(a.cfg)
	case PageVersion:
		a.editorPage = NewVersionPage(a.cfg)
	case PageImport:
		a.importPage = NewImportPage("")
	case PageDeploy:
		a.deployPage = NewDeployPage(nil)
	}
}

func (a *App) goBack() {
	if len(a.pageStack) > 0 {
		a.currentPage = a.pageStack[len(a.pageStack)-1]
		a.pageStack = a.pageStack[:len(a.pageStack)-1]
	}
}

// View implements tea.Model.
func (a *App) View() string {
	var b strings.Builder

	// Title bar.
	title := titleStyle.Render(fmt.Sprintf("spinctl v%s", a.version))
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(a.width, 40)))
	b.WriteString("\n")

	// Current page content.
	switch a.currentPage {
	case PageHome:
		if a.homePage != nil {
			b.WriteString(a.homePage.View())
		}
	case PageServices:
		if a.servicesPage != nil {
			b.WriteString(a.servicesPage.View())
		}
	case PageProviders, PageSecurity, PageFeatures, PageVersion, PageEditor:
		if a.editorPage != nil {
			b.WriteString(a.editorPage.View())
		}
	case PageImport:
		if a.importPage != nil {
			b.WriteString(a.importPage.View())
		}
	case PageDeploy:
		if a.deployPage != nil {
			b.WriteString(a.deployPage.View())
		}
	}

	b.WriteString("\n")

	// Status bar.
	a.statusBar.SetModified(a.dirty)
	b.WriteString(a.statusBar.View("q: quit  ?: help"))

	return b.String()
}
