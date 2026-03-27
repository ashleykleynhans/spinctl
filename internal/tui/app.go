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
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

// page is the interface each TUI page must implement.
type page interface {
	Update(msg tea.Msg) (page, tea.Cmd)
	View() string
}

// saveResultMsg is sent after a save attempt.
type saveResultMsg struct {
	err error
}

// goBackMsg signals the app to navigate back to the previous page.
type goBackMsg struct{}

// App is the root bubbletea model and page router.
type App struct {
	cfg          *config.SpinctlConfig
	configPath   string
	version      string
	currentPage  PageID
	pageStack    []PageID
	homePage     page
	servicesPage page
	editorPage   page
	importPage   page
	deployPage   page
	dirty        bool
	width        int
	height       int
	statusBar    *components.StatusBar
	confirmQuit  bool
	saveMessage  string
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

	case importDoneMsg:
		// After import completes, reload config and refresh home page.
		if msg.err == nil && msg.result != nil && msg.result.Config != nil {
			a.cfg = msg.result.Config
			a.homePage = NewHomePage(a.cfg)
			a.dirty = false
		}
		// Still delegate to import page so it can show the result.

	case goBackMsg:
		a.goBack()
		return a, nil

	case saveResultMsg:
		if msg.err != nil {
			a.saveMessage = warnStyle.Render(fmt.Sprintf("Save failed: %s", msg.err))
		} else {
			a.dirty = false
			a.saveMessage = successStyle.Render("Config saved")
		}
		return a, nil

	case tea.KeyMsg:
		// Clear save message on any key press.
		a.saveMessage = ""

		// Handle quit confirmation.
		if a.confirmQuit {
			switch msg.String() {
			case "y":
				return a, tea.Quit
			case "n":
				a.confirmQuit = false
				return a, nil
			case "s":
				// Save then quit.
				a.confirmQuit = false
				return a, tea.Sequence(a.saveConfig(), func() tea.Msg {
					return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
				})
			}
			return a, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if a.dirty {
				a.confirmQuit = true
				return a, nil
			}
			return a, tea.Quit
		case "s":
			if a.currentPage == PageHome {
				return a, a.saveConfig()
			}
		case "esc":
			// Only handle esc at app level for simple pages.
			// Pages with internal navigation (services, providers,
			// security, editor) handle esc themselves.
			switch a.currentPage {
			case PageFeatures, PageImport, PageDeploy:
				a.goBack()
				return a, nil
			case PageHome:
				// Do nothing on home page.
			default:
				// Let the page handle esc first.
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
			a.markDirty()
		}
	case PageProviders, PageSecurity, PageFeatures, PageVersion, PageEditor:
		if a.editorPage != nil {
			var updated page
			updated, cmd = a.editorPage.Update(msg)
			a.editorPage = updated
			a.markDirty()
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

func (a *App) markDirty() {
	a.dirty = true
}

func (a *App) saveConfig() tea.Cmd {
	return func() tea.Msg {
		err := config.SaveToFile(a.cfg, a.configPath)
		return saveResultMsg{err: err}
	}
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

	// Quit confirmation overlay.
	if a.confirmQuit {
		b.WriteString("\n")
		b.WriteString(warnStyle.Render("  You have unsaved changes."))
		b.WriteString("\n\n")
		b.WriteString("  s: save and quit  y: quit without saving  n: cancel\n")
		return b.String()
	}

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

	// Save message.
	if a.saveMessage != "" {
		b.WriteString("  " + a.saveMessage + "\n")
	}

	// Status bar.
	a.statusBar.SetModified(a.dirty)
	hints := "s: save  q: quit  ?: help"
	if a.currentPage != PageHome {
		hints = "esc: back  s: save  q: quit"
	}
	b.WriteString(a.statusBar.View(hints))

	return b.String()
}
