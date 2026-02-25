package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	commandpalette "ldt-toolkit-cli/internal/command_palette"
	dataprep "ldt-toolkit-cli/internal/screens/data_preparation"
	datapp "ldt-toolkit-cli/internal/screens/data_preprocessing"
	"ldt-toolkit-cli/internal/screens/home"
	machinelearning "ldt-toolkit-cli/internal/screens/machine_learning"
	anim "ldt-toolkit-cli/internal/shared/asciimotion"
	"ldt-toolkit-cli/internal/shared/common"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	defaultUvBinary       = "uv"
	pythonEntryPoint      = "ldt-toolkit"
	bridgeRunnerCommand   = "ldt-bridge"
	bridgeRunnerModule    = "ldt.bridge.runner"
	installationHelpHint  = "Run from a project directory containing pyproject.toml with `uv add ldt-toolkit`, then retry."
	bridgeHintDelay       = 9 * time.Second
	bridgeHintRotateEvery = 4 * time.Second
	homeTabLabel          = "Home"
	contentPanelMinWidth  = 64
	contentPanelMinHeight = 12
	listPanelMaxWidth     = 100
	stagePanelWidth       = 108
)

var (
	inNavigatorSession  bool
	ansiEscapePattern   = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	activeBridgeRuntime bridgeRuntime
	bridgeWaitHints     = []string{
		"Please wait, Python is slow the first time it imports required packages...",
		"Please wait, warming up the runtime...",
		"Please wait, still loading dependencies...",
		"Please wait, nearly there!",
		"Please wait, finishing up...",
		"Please wait, just a moment longer...",
	}
)

type commandDef = model.CommandDef
type parsedHelp = model.ParsedHelp

type listItem struct {
	command commandDef
}

func (i listItem) Title() string       { return model.CommandLabel(i.command) }
func (i listItem) Description() string { return i.command.Description }
func (i listItem) FilterValue() string { return i.command.Name }

type stageState struct {
	root    string
	path    []string
	node    *parsedHelp
	loadErr string
	list    list.Model
}

type loadNodeMsg struct {
	stage string
	path  []string
	node  *parsedHelp
	err   error
}

type appModel struct {
	width  int
	height int

	activeTab int
	tabs      []string

	loader *components.HelpLoader

	spinner spinner.Model
	loading bool

	stages map[string]*stageState

	homeScreen home.Screen
	homeShine  *components.ShinyTitle

	palette commandpalette.Model

	pendingRun []string

	headerAnimation anim.Model

	stageShines map[string]*components.ShinyTitle
}

type triggerActivePanelTitleMsg struct{}

type bridgeRuntime struct {
	useUv        bool
	uvBinary     string
	pythonBinary string
	workingDir   string
	useBridgeCmd bool
}

var stageCommands = []commandDef{
	{
		Name:        "data_preparation",
		DisplayName: "Data Preparation",
		Description: "(1-A) Prepare raw cohort longitudinal data or generate synthetic longitudinal data.",
	},
	{
		Name:        "data_preprocessing",
		DisplayName: "Data Preprocessing",
		Description: "(2-B) Build trajectories, improve data quality, and manipulate datasets for modelling.",
	},
	{
		Name:        "machine_learning",
		DisplayName: "Machine Learning",
		Description: "(3-C) Run longitudinal and standard machine learning workflows.",
	},
}

func Run(args []string) error {
	runtime, err := resolveBridgeRuntime()
	if err != nil {
		return err
	}
	activeBridgeRuntime = runtime

	if err := dataprep.Configure(); err != nil {
		return err
	}
	if err := datapp.Configure(); err != nil {
		return err
	}
	if err := machinelearning.Configure(); err != nil {
		return err
	}
	homeScreen, err := configureHomeContent()
	if err != nil {
		return err
	}
	dataprep.ConfigureRuntime(
		executeBridgeOperation,
		func() bool { return inNavigatorSession },
	)
	datapp.ConfigureRuntime(
		executeBridgeOperation,
		func() bool { return inNavigatorSession },
	)
	machinelearning.ConfigureRuntime(
		executeBridgeOperation,
		func() bool { return inNavigatorSession },
	)

	if len(args) > 0 {
		return errors.New("interactive mode only: run `ldt` without arguments")
	}
	return runInteractiveNavigator(homeScreen)
}

func runInteractiveNavigator(homeScreen home.Screen) error {
	inNavigatorSession = true
	defer func() {
		inNavigatorSession = false
	}()

	appState, err := newAppModel(newHelpLoader(), homeScreen)
	if err != nil {
		return err
	}

	for {
		program := tea.NewProgram(appState, tea.WithAltScreen())
		finalModel, err := program.Run()
		if err != nil {
			return err
		}

		resolved, ok := finalModel.(appModel)
		if !ok {
			return errors.New("unexpected Bubble Tea model type")
		}
		appState = resolved

		if len(appState.pendingRun) == 0 {
			return nil
		}

		runArgs := model.ClonePath(appState.pendingRun)
		handled, err := runBridgeOperationForPath(runArgs)
		if err != nil {
			return err
		}
		if handled {
			appState.pendingRun = nil
			appState.onBridgeActionComplete(runArgs)
			continue
		}
		return fmt.Errorf("this action is not yet available in the Go interactive runtime: %s", strings.Join(runArgs, " "))
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		m.headerAnimation.Init(),
		func() tea.Msg { return triggerActivePanelTitleMsg{} },
	)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	titleCmd := m.updatePanelTitleAnimations(msg)

	if _, ok := msg.(anim.TickMsg); ok {
		updated, cmd := m.headerAnimation.Update(msg)
		if next, ok := updated.(anim.Model); ok {
			m.headerAnimation = next
		}
		return m, tea.Batch(cmd, titleCmd)
	}

	switch message := msg.(type) {
	case triggerActivePanelTitleMsg:
		return m, tea.Batch(m.resetActivePanelTitle(), titleCmd)

	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		m.resizeLists()
		return m, tea.Batch(m.ensureActiveStageLoaded(), titleCmd)

	case loadNodeMsg:
		m.loading = false
		stage := m.stages[message.stage]
		if stage == nil {
			return m, titleCmd
		}
		if model.JoinPath(stage.path) != model.JoinPath(message.path) {
			return m, titleCmd
		}
		if message.err != nil {
			stage.node = nil
			stage.loadErr = message.err.Error()
			stage.list.SetItems(nil)
			return m, titleCmd
		}

		stage.node = message.node
		stage.loadErr = ""
		stage.list.Title = ""
		stage.list.SetItems(commandsToItems(stage.node.Commands))
		m.resizeLists()
		if (isDataPreparationStage(message.stage) || isDataPreprocessingStage(message.stage) || isMachineLearningStage(message.stage)) &&
			message.node != nil && len(message.node.Commands) == 0 {
			m.pendingRun = model.ClonePath(message.path)
			return m, tea.Batch(tea.Quit, titleCmd)
		}
		return m, titleCmd

	case spinner.TickMsg:
		if !m.loading {
			return m, titleCmd
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, titleCmd)

	case tea.KeyMsg:
		if m.palette.IsOpen() {
			next, cmd := m.updatePalette(message)
			return next, tea.Batch(cmd, titleCmd)
		}
		switch message.String() {
		case "ctrl+c", "q":
			m.pendingRun = nil
			return m, tea.Batch(tea.Quit, titleCmd)
		case ":", "ctrl+p":
			return m, tea.Batch(m.palette.Open(), titleCmd)
		case "left", "shift+tab":
			return m, tea.Batch(m.switchTab(-1), titleCmd)
		case "right", "tab":
			return m, tea.Batch(m.switchTab(1), titleCmd)
		}
	}

	if m.palette.IsOpen() {
		next, cmd := m.updatePalette(msg)
		return next, tea.Batch(cmd, titleCmd)
	}
	if m.activeTab == 0 {
		next, cmd := m.updateHome(msg)
		return next, tea.Batch(cmd, titleCmd)
	}
	next, cmd := m.updateStage(msg)
	return next, tea.Batch(cmd, titleCmd)
}

func (m appModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading ldt..."
	}

	header := components.ApplyLeftLayoutMargin(renderInteractiveHeader(m.width, m.headerAnimation.View()))
	tabs := components.ApplyLeftLayoutMargin(components.RenderNavigationTabs(m.tabDisplayLabels(), m.activeTab))
	content := components.ApplyLeftLayoutMargin(m.renderContent())
	footer := components.ApplyLeftLayoutMargin(components.RenderNavigatorFooter(m.activeTab))
	if m.palette.IsOpen() {
		tabs = softenForPaletteOverlay(tabs)
		content = softenForPaletteOverlay(content)
		footer = softenForPaletteOverlay(footer)
	}

	rows := []string{header, tabs, content}
	if m.palette.IsOpen() {
		rows = append(rows, components.ApplyLeftLayoutMargin(m.renderPalette()))
	}
	rows = append(rows, footer)
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *appModel) switchTab(delta int) tea.Cmd {
	tabCount := len(m.tabs)
	if tabCount == 0 {
		return nil
	}
	m.activeTab = (m.activeTab + delta + tabCount) % tabCount
	return tea.Batch(m.ensureActiveStageLoaded(), m.resetActivePanelTitle())
}

func (m *appModel) updatePanelTitleAnimations(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0, 1+len(m.stageShines))

	if m.homeShine != nil {
		_, cmd := m.homeShine.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	for _, shine := range m.stageShines {
		if shine == nil {
			continue
		}
		_, cmd := shine.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

func (m *appModel) resetActivePanelTitle() tea.Cmd {
	if m.activeTab == 0 {
		if m.homeShine == nil {
			title := components.NewShinyTitle("Home")
			m.homeShine = &title
		}
		m.homeShine.SetText("Home")
		return m.homeShine.Reset()
	}

	stageName := m.tabs[m.activeTab]
	if strings.TrimSpace(stageName) == "" {
		return nil
	}

	shine := m.stageShines[stageName]
	if shine == nil {
		title := components.NewShinyTitle(stageDisplayName(stageName))
		shine = &title
		if m.stageShines == nil {
			m.stageShines = map[string]*components.ShinyTitle{}
		}
		m.stageShines[stageName] = shine
	}
	shine.SetText(m.currentStageHeading(stageName))
	return shine.Reset()
}

func (m appModel) renderedHomeHeading() string {
	if m.homeShine == nil {
		return theme.App.SubSubtitleStyle().Render("Home")
	}
	return m.homeShine.ViewWithText(theme.App.SubSubtitleStyle(), "Home")
}

func (m appModel) renderedStageHeading(stageName string, heading string) string {
	shine := m.stageShines[stageName]
	if shine == nil {
		return theme.App.SubSubtitleStyle().Render(strings.TrimSpace(heading))
	}
	return shine.ViewWithText(theme.App.SubSubtitleStyle(), heading)
}

func (m appModel) currentStageHeading(stageName string) string {
	stage := m.stages[stageName]
	if stage == nil {
		return stageDisplayName(stageName)
	}
	return stageHeadingText(stageName, stage.path)
}

func (m *appModel) ensureActiveStageLoaded() tea.Cmd {
	if m.activeTab == 0 || m.loading {
		return nil
	}
	stageName := m.tabs[m.activeTab]
	stage := m.stages[stageName]
	if stage == nil {
		return nil
	}
	if stage.node != nil || stage.loadErr != "" {
		return nil
	}
	m.loading = true
	return tea.Batch(loadNodeCmd(stageName, stage.path, m.loader), m.spinner.Tick)
}

func (m *appModel) onBridgeActionComplete(path []string) {
	if canonical, handled := dataprep.CanonicalPath(path); handled {
		m.resetStageAfterAction("data_preparation", canonical)
		return
	}
	if canonical, handled := datapp.CanonicalPath(path); handled {
		m.resetStageAfterAction("data_preprocessing", canonical)
		return
	}
	if canonical, handled := machinelearning.CanonicalPath(path); handled {
		m.resetStageAfterAction("machine_learning", canonical)
	}
}

func (m *appModel) resetStageAfterAction(stageName string, canonical []string) {
	stage := m.stages[stageName]
	if stage == nil {
		return
	}

	targetPath := model.ClonePath(canonical)
	if len(targetPath) > 1 {
		targetPath = targetPath[:len(targetPath)-1]
	}
	if len(targetPath) == 0 {
		targetPath = []string{stageName}
	}

	stage.path = targetPath
	stage.node = nil
	stage.loadErr = ""
	stage.list.SetItems(nil)
	m.loading = false

	for index, tab := range m.tabs {
		if strings.EqualFold(tab, stageName) {
			m.activeTab = index
			break
		}
	}
}

func (m *appModel) exitCurrentTool(stageName string, stage *stageState) bool {
	if isDataPreparationStage(stageName) {
		canonical, handled := dataprep.CanonicalPath(stage.path)
		if handled && len(canonical) >= 3 && canonical[1] == "tools" {
			stage.path = model.ClonePath(canonical[:2])
			return true
		}
	}
	if isDataPreprocessingStage(stageName) {
		canonical, handled := datapp.CanonicalPath(stage.path)
		if handled && len(canonical) >= 3 && canonical[1] == "tools" {
			stage.path = model.ClonePath(canonical[:2])
			return true
		}
		if handled && len(canonical) >= 3 && canonical[1] == "presets" {
			stage.path = model.ClonePath(canonical[:2])
			return true
		}
	}
	if isMachineLearningStage(stageName) {
		canonical, handled := machinelearning.CanonicalPath(stage.path)
		if handled && len(canonical) >= 3 && canonical[1] == "tools" {
			stage.path = model.ClonePath(canonical[:2])
			return true
		}
		if handled && len(canonical) >= 3 && canonical[1] == "presets" {
			stage.path = model.ClonePath(canonical[:2])
			return true
		}
	}

	if len(stage.path) > 1 {
		stage.path = []string{stage.root}
		return true
	}
	return false
}

func (m appModel) updateHome(msg tea.Msg) (tea.Model, tea.Cmd) {
	if pending := m.homeScreen.Update(msg); len(pending) > 0 {
		m.pendingRun = pending
		return m, tea.Quit
	}
	return m, nil
}

func (m appModel) updateStage(msg tea.Msg) (tea.Model, tea.Cmd) {
	stageName := m.tabs[m.activeTab]
	stage := m.stages[stageName]
	if stage == nil {
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "b", "backspace", "esc":
			if m.loading {
				return m, nil
			}
			if len(stage.path) > 1 {
				stage.path = model.ClonePath(stage.path[:len(stage.path)-1])
				stage.node = nil
				stage.loadErr = ""
				stage.list.SetItems(nil)
				m.loading = true
				return m, tea.Batch(loadNodeCmd(stageName, stage.path, m.loader), m.spinner.Tick)
			}
			m.activeTab = 0
			return m, (&m).resetActivePanelTitle()

		case "l":
			if m.loading {
				return m, nil
			}
			stage.node = nil
			stage.loadErr = ""
			stage.list.SetItems(nil)
			m.loading = true
			return m, tea.Batch(loadNodeCmd(stageName, stage.path, m.loader), m.spinner.Tick)

		case "x":
			if m.loading {
				return m, nil
			}
			if !m.exitCurrentTool(stageName, stage) {
				return m, nil
			}
			stage.node = nil
			stage.loadErr = ""
			stage.list.SetItems(nil)
			m.loading = true
			return m, tea.Batch(loadNodeCmd(stageName, stage.path, m.loader), m.spinner.Tick)

		case "enter":
			if m.loading {
				return m, nil
			}
			if stage.node == nil {
				m.loading = true
				return m, tea.Batch(loadNodeCmd(stageName, stage.path, m.loader), m.spinner.Tick)
			}
			if len(stage.node.Commands) == 0 {
				m.pendingRun = model.ClonePath(stage.path)
				return m, tea.Quit
			}
			selectedItem, ok := stage.list.SelectedItem().(listItem)
			if !ok {
				return m, nil
			}
			stage.path = append(model.ClonePath(stage.path), selectedItem.command.Name)
			stage.node = nil
			stage.loadErr = ""
			stage.list.SetItems(nil)
			m.loading = true
			return m, tea.Batch(loadNodeCmd(stageName, stage.path, m.loader), m.spinner.Tick)
		}
	}

	if m.loading || stage.node == nil || len(stage.node.Commands) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	stage.list, cmd = stage.list.Update(msg)
	return m, cmd
}

func (m appModel) updatePalette(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd, resolved := m.palette.Update(msg)
	if resolved == nil {
		return m, cmd
	}
	m.pendingRun = model.ClonePath(resolved.Args)
	return m, tea.Quit
}

func (m *appModel) resizeLists() {
	panelWidth := stagePanelWidthForTerminal(m.width)
	panelInnerWidth := model.IntMax(contentPanelMinWidth, panelWidth-6)
	contentWidth := model.IntMin(listPanelMaxWidth, panelInnerWidth)
	stageListHeight := stageListHeightForTerminal(m.height)

	for _, stage := range m.stages {
		stage.list.SetSize(contentWidth, stageListHeight)
	}
	m.palette.Resize(m.width)
}

func (m appModel) renderContent() string {
	if m.activeTab == 0 {
		panelWidth := stagePanelWidthForTerminal(m.width)
		panelBodyWidth := model.IntMax(contentPanelMinWidth, panelWidth-6)
		panelHeight := homePanelHeightForTerminal(m.height)
		homeBody := m.homeScreen.Render(panelBodyWidth, m.renderedHomeHeading())
		panelHeight = model.IntMax(panelHeight, lipgloss.Height(homeBody))
		return theme.App.PanelStyle().
			Width(panelWidth).
			Height(panelHeight).
			Render(homeBody)
	}

	stageName := m.tabs[m.activeTab]
	stage := m.stages[stageName]
	if stage == nil {
		return theme.App.PanelStyle().Render("Stage not available.")
	}

	summary := ""
	if stage.node != nil {
		summary = stage.node.Summary
	}
	usage := ""
	if stage.node != nil && stageShouldDisplayUsage(stageName) && strings.TrimSpace(stage.node.Usage) != "" {
		usage = stage.node.Usage
	}

	nextPreview := ""
	if stage.node != nil && len(stage.node.Commands) > 0 {
		if selected, ok := stage.list.SelectedItem().(listItem); ok {
			nextTarget := strings.Join(append(model.ClonePath(stage.path), selected.command.Name), " > ")
			if isDataPreparationStage(stageName) {
				nextTarget = model.CommandLabel(selected.command)
			}
			nextPreview = fmt.Sprintf("Open next: %s", nextTarget)
		}
	}

	listView := ""
	if stage.node != nil && len(stage.node.Commands) > 0 {
		listView = stage.list.View()
	}

	panelWidth := stagePanelWidthForTerminal(m.width)
	panelBodyWidth := model.IntMax(contentPanelMinWidth, panelWidth-6)
	heading := stageHeadingText(stageName, stage.path)

	return components.RenderStagePanel(components.StagePanelInput{
		Heading:          heading,
		HeadingRendered:  m.renderedStageHeading(stageName, heading),
		PanelWidth:       panelWidth,
		BodyWidth:        panelBodyWidth,
		TruncateOverflow: true,

		Loading:          m.loading,
		LoadingIndicator: m.spinner.View(),
		LoadingMessage:   loadingStatusMessage(stageName),

		LoadError: stage.loadErr,
		LoadHint:  "Could not load this node. Press l to retry or b to go back.",

		NodeMissingHint: "Press l to load this stage node.",

		Summary: summary,
		Usage:   usage,

		IsDataPreparationLeaf: (isDataPreparationStage(stageName) || isDataPreprocessingStage(stageName) || isMachineLearningStage(stageName)) &&
			stage.node != nil && len(stage.node.Commands) == 0,
		RunTargetPath: strings.Join(stage.path, " > "),
		RunHint:       "Press Enter to run this action.",

		ListView:    listView,
		NextPreview: nextPreview,
	})
}

func stagePanelWidthForTerminal(terminalWidth int) int {
	maxAllowed := model.IntMax(contentPanelMinWidth, terminalWidth-6)
	if maxAllowed < stagePanelWidth {
		return maxAllowed
	}
	return stagePanelWidth
}

func stageListHeightForTerminal(terminalHeight int) int {
	return model.IntMax(contentPanelMinHeight, model.IntMin(16, terminalHeight-30))
}

func homePanelHeightForTerminal(terminalHeight int) int {
	// Align the Home panel visual height with stage panels (heading + list + preview rows).
	return stageListHeightForTerminal(terminalHeight) + 5
}

func (m appModel) renderPalette() string {
	return m.palette.View()
}

func (m appModel) tabDisplayLabels() []string {
	labels := make([]string, len(m.tabs))
	for index, tab := range m.tabs {
		if tab == homeTabLabel {
			labels[index] = homeTabLabel
			continue
		}
		labels[index] = stageDisplayName(tab)
	}
	return labels
}

func newAppModel(loader *components.HelpLoader, homeScreen home.Screen) (appModel, error) {
	tabs := []string{homeTabLabel}
	for _, command := range stageCommands {
		tabs = append(tabs, command.Name)
	}

	delegate := components.StyledSubActionDelegate()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = theme.App.AccentStyle()

	palette, err := commandpalette.New(findProjectRoot)
	if err != nil {
		return appModel{}, fmt.Errorf("failed to initialize command palette: %w", err)
	}

	stages := make(map[string]*stageState, len(stageCommands))
	stageShines := make(map[string]*components.ShinyTitle, len(stageCommands))
	for _, command := range stageCommands {
		stages[command.Name] = &stageState{
			root: command.Name,
			path: []string{command.Name},
			list: components.NewSubActionList(model.CommandLabel(command), nil, delegate, 80, 12),
		}
		title := components.NewShinyTitle(stageDisplayName(command.Name))
		stageShines[command.Name] = &title
	}

	homeShine := components.NewShinyTitle("Home")

	headerAnimation := anim.New(anim.Config{
		AutoPlay:          true,
		Loop:              false,
		HasDarkBackground: true,
		Width:             80,
		Height:            24,
	})

	return appModel{
		activeTab:   0,
		tabs:        tabs,
		loader:      loader,
		spinner:     sp,
		stages:      stages,
		homeScreen:  homeScreen,
		homeShine:   &homeShine,
		palette:     palette,
		stageShines: stageShines,

		headerAnimation: headerAnimation,
	}, nil
}

func newHelpLoader() *components.HelpLoader {
	helpUvBinary := defaultUvBinary
	if activeBridgeRuntime.useUv && strings.TrimSpace(activeBridgeRuntime.uvBinary) != "" {
		helpUvBinary = activeBridgeRuntime.uvBinary
	}
	return components.NewHelpLoader(components.HelpLoaderConfig{
		UvBinary:          helpUvBinary,
		EntryPoint:        pythonEntryPoint,
		Environment:       os.Environ(),
		ProjectRootFinder: findProjectRoot,
		CustomNodeLoader: func(path []string) (*model.ParsedHelp, bool, error) {
			if node, handled, err := dataprep.LoadNode(path); handled {
				return node, true, err
			}
			if node, handled, err := datapp.LoadNode(path); handled {
				return node, true, err
			}
			return machinelearning.LoadNode(path)
		},
	})
}

func loadNodeCmd(stage string, path []string, loader *components.HelpLoader) tea.Cmd {
	safePath := model.ClonePath(path)
	return func() tea.Msg {
		node, err := loader.Load(safePath)
		return loadNodeMsg{
			stage: stage,
			path:  safePath,
			node:  node,
			err:   err,
		}
	}
}

func commandsToItems(commands []commandDef) []list.Item {
	items := make([]list.Item, 0, len(commands))
	for _, command := range commands {
		items = append(items, listItem{command: command})
	}
	return items
}

func renderInteractiveHeader(width int, banner string) string {
	return components.RenderInteractiveHeaderWithBanner(width, banner)
}

func softenForPaletteOverlay(content string) string {
	plain := ansiEscapePattern.ReplaceAllString(content, "")
	return lipgloss.NewStyle().
		Foreground(theme.App.Color("#8E8895")).
		Background(theme.App.Color("#1C1C23")).
		Render(plain)
}

func stageHeadingText(stageName string, stagePath []string) string {
	if isDataPreparationStage(stageName) {
		return dataprep.Heading(stagePath)
	}
	if isDataPreprocessingStage(stageName) {
		return datapp.Heading(stagePath)
	}
	if isMachineLearningStage(stageName) {
		return machinelearning.Heading(stagePath)
	}
	parts := make([]string, 0, len(stagePath))
	for _, token := range stagePath {
		parts = append(parts, stagePathTokenLabel(token))
	}
	return fmt.Sprintf("Path: %s", strings.Join(parts, " > "))
}

func stageDisplayName(name string) string {
	for _, stage := range stageCommands {
		if stage.Name == name {
			return model.CommandLabel(stage)
		}
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "Unnamed"
	}
	return trimmed
}

func stagePathTokenLabel(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return "Unnamed"
	}
	for _, stage := range stageCommands {
		if stage.Name == trimmed {
			return model.CommandLabel(stage)
		}
	}
	return trimmed
}

func stageShouldDisplayUsage(stageName string) bool {
	return !isDataPreparationStage(stageName) &&
		!isDataPreprocessingStage(stageName) &&
		!isMachineLearningStage(stageName)
}

func isDataPreparationStage(stageName string) bool {
	return strings.EqualFold(strings.TrimSpace(stageName), "data_preparation")
}

func isDataPreprocessingStage(stageName string) bool {
	return strings.EqualFold(strings.TrimSpace(stageName), "data_preprocessing")
}

func isMachineLearningStage(stageName string) bool {
	return strings.EqualFold(strings.TrimSpace(stageName), "machine_learning")
}

func loadingStatusMessage(stageName string) string {
	if strings.EqualFold(stageName, "data_preparation") {
		return "Loading data preparation catalog..."
	}
	if strings.EqualFold(stageName, "data_preprocessing") {
		return "Loading data preprocessing catalog..."
	}
	if strings.EqualFold(stageName, "machine_learning") {
		return "Loading machine learning catalog..."
	}
	return "Loading command map..."
}

type bridgeErrorEnvelope struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type bridgeEnvelope struct {
	OK     bool                 `json:"ok"`
	Result map[string]any       `json:"result"`
	Error  *bridgeErrorEnvelope `json:"error"`
}

type bridgeExecutionDoneMsg struct {
	runErr error
	stdout string
	stderr string
}

type bridgeExecutionModel struct {
	operation string
	stopwatch stopwatch.Model
	doneCh    <-chan bridgeExecutionDoneMsg
	result    *bridgeExecutionDoneMsg
}

func runBridgeOperationForPath(path []string) (bool, error) {
	if handled, err := dataprep.HandleBridgePath(path); handled || err != nil {
		return handled, err
	}
	if handled, err := datapp.HandleBridgePath(path); handled || err != nil {
		return handled, err
	}
	if handled, err := machinelearning.HandleBridgePath(path); handled || err != nil {
		return handled, err
	}

	switch model.JoinPath(path) {
	case model.JoinPath([]string{"list_authors"}):
		return true, home.RunConfiguredAuthors()

	case model.JoinPath([]string{"list_inspiration_from"}):
		return true, home.RunConfiguredInspirations()

	case model.JoinPath([]string{"toolkit_tutorial"}):
		return true, home.RunConfiguredTutorial()
	}

	return false, nil
}

func executeBridgeOperation(operation string, params map[string]any) (map[string]any, error) {
	if err := ensureBridgeRuntimeReady(); err != nil {
		return nil, err
	}

	payload := map[string]any{
		"operation": operation,
		"params":    params,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bridge payload: %w", err)
	}

	cmd := activeBridgeRuntime.bridgeCommand()
	cmd.Dir = activeBridgeRuntime.workingDir
	cmd.Stdin = bytes.NewReader(jsonPayload)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if shouldShowBridgeExecutionStatus() {
		return executeBridgeOperationWithStopwatch(cmd, operation)
	}

	runErr := cmd.Run()
	return parseBridgeExecutionResult(runErr, stdout.String(), stderr.String())
}

func shouldShowBridgeExecutionStatus() bool {
	return inNavigatorSession
}

func executeBridgeOperationWithStopwatch(
	cmd *exec.Cmd,
	operation string,
) (map[string]any, error) {
	doneCh := make(chan bridgeExecutionDoneMsg, 1)
	go func() {
		runErr := cmd.Run()
		stdoutText := ""
		stderrText := ""
		if buffer, ok := cmd.Stdout.(*bytes.Buffer); ok {
			stdoutText = buffer.String()
		}
		if buffer, ok := cmd.Stderr.(*bytes.Buffer); ok {
			stderrText = buffer.String()
		}
		doneCh <- bridgeExecutionDoneMsg{
			runErr: runErr,
			stdout: stdoutText,
			stderr: stderrText,
		}
	}()

	program := tea.NewProgram(newBridgeExecutionModel(operation, doneCh))
	finalModel, runErr := program.Run()
	if runErr != nil {
		done := <-doneCh
		return parseBridgeExecutionResult(done.runErr, done.stdout, done.stderr)
	}

	model, ok := finalModel.(*bridgeExecutionModel)
	if !ok || model == nil || model.result == nil {
		done := <-doneCh
		return parseBridgeExecutionResult(done.runErr, done.stdout, done.stderr)
	}
	return parseBridgeExecutionResult(
		model.result.runErr,
		model.result.stdout,
		model.result.stderr,
	)
}

func parseBridgeExecutionResult(
	runErr error,
	stdoutText string,
	stderrText string,
) (map[string]any, error) {
	rawOutput := strings.TrimSpace(stdoutText)
	if rawOutput == "" {
		if runErr != nil {
			return nil, fmt.Errorf("bridge execution failed: %w\n%s", runErr, strings.TrimSpace(stderrText))
		}
		return nil, errors.New("bridge execution returned no output")
	}

	var envelope bridgeEnvelope
	if err := json.Unmarshal([]byte(rawOutput), &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode bridge response: %w\n%s", err, rawOutput)
	}

	if !envelope.OK {
		if envelope.Error != nil {
			return nil, fmt.Errorf("%s: %s", envelope.Error.Type, formatBridgeErrorMessage(envelope.Error.Message))
		}
		return nil, errors.New("bridge returned an unknown error")
	}

	if runErr != nil {
		return nil, fmt.Errorf("bridge process failed: %w", runErr)
	}

	if envelope.Result == nil {
		return map[string]any{}, nil
	}
	return envelope.Result, nil
}

func formatBridgeErrorMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return message
	}

	// Keep this generic: the toolkit now pins its own compatibility deps.
	if strings.Contains(trimmed, "No module named 'pkg_resources'") ||
		strings.Contains(trimmed, "No module named \"pkg_resources\"") {
		return trimmed + "\nHint: sync project dependencies with `uv sync`, then retry."
	}

	return trimmed
}

func newBridgeExecutionModel(
	operation string,
	doneCh <-chan bridgeExecutionDoneMsg,
) *bridgeExecutionModel {
	sw := stopwatch.NewWithInterval(time.Second)
	return &bridgeExecutionModel{
		operation: operation,
		stopwatch: sw,
		doneCh:    doneCh,
	}
}

func (m *bridgeExecutionModel) Init() tea.Cmd {
	return tea.Batch(
		m.stopwatch.Init(),
		waitBridgeExecutionCmd(m.doneCh),
	)
}

func (m *bridgeExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case bridgeExecutionDoneMsg:
		m.result = &typed
		return m, tea.Quit
	}

	var stopwatchCmd tea.Cmd
	m.stopwatch, stopwatchCmd = m.stopwatch.Update(msg)
	return m, stopwatchCmd
}

func (m *bridgeExecutionModel) View() string {
	action := strings.ReplaceAll(strings.TrimSpace(m.operation), ".", " / ")
	if action == "" {
		action = "bridge operation"
	}
	waitMessage := bridgeWaitStatusMessage(m.stopwatch.Elapsed())
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		theme.App.SubtitleStyle().Render(fmt.Sprintf("Running %s", action)),
		theme.App.TextStyle().Render(fmt.Sprintf("Elapsed: %s", m.stopwatch.View())),
		theme.App.MutedTextStyle().Render(waitMessage),
		"",
	)
	return components.ApplyLeftLayoutMargin(content)
}

func bridgeWaitStatusMessage(elapsed time.Duration) string {
	const defaultMessage = "Please wait..."

	if elapsed <= bridgeHintDelay || len(bridgeWaitHints) == 0 {
		return defaultMessage
	}

	offset := elapsed - bridgeHintDelay
	if offset < 0 {
		return defaultMessage
	}

	index := int(offset / bridgeHintRotateEvery)
	if index < 0 {
		return defaultMessage
	}
	return bridgeWaitHints[index%len(bridgeWaitHints)]
}

func waitBridgeExecutionCmd(doneCh <-chan bridgeExecutionDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return <-doneCh
	}
}

func ensureBridgeRuntimeReady() error {
	if activeBridgeRuntime.useUv || strings.TrimSpace(activeBridgeRuntime.pythonBinary) != "" {
		return nil
	}
	runtime, err := resolveBridgeRuntime()
	if err != nil {
		return err
	}
	activeBridgeRuntime = runtime
	return nil
}

func resolveBridgeRuntime() (bridgeRuntime, error) {
	workingDir, err := findRuntimeWorkingDir()
	if err != nil {
		return bridgeRuntime{}, err
	}

	uvPath, uvLookupErr := exec.LookPath(defaultUvBinary)
	uvExists := uvLookupErr == nil
	pythonPath, pythonExists := findPythonBinary()

	if !uvExists && !pythonExists {
		return bridgeRuntime{}, fmt.Errorf(
			"`uv` or `python` was not found on PATH. %s",
			installationHelpHint,
		)
	}

	var uvErr error
	if uvExists {
		bridgeCmdRuntime := bridgeRuntime{
			useUv:        true,
			uvBinary:     uvPath,
			pythonBinary: "python",
			workingDir:   workingDir,
			useBridgeCmd: true,
		}
		if err := verifyBridgeCommand(bridgeCmdRuntime); err == nil {
			return bridgeCmdRuntime, nil
		}
		uvErr = err
	}

	var pythonErr error
	if pythonExists {
		runtime := bridgeRuntime{
			useUv:        false,
			pythonBinary: pythonPath,
			workingDir:   workingDir,
		}
		if err := verifyPythonBridge(runtime); err == nil {
			return runtime, nil
		}
		pythonErr = err
	}

	if uvErr != nil && pythonErr != nil {
		return bridgeRuntime{}, fmt.Errorf(
			"bridge runtime checks failed.\n- uv: %v\n- python: %v",
			uvErr,
			pythonErr,
		)
	}
	if uvErr != nil {
		return bridgeRuntime{}, uvErr
	}
	if pythonErr != nil {
		return bridgeRuntime{}, pythonErr
	}
	return bridgeRuntime{}, fmt.Errorf("no usable bridge runtime found. %s", installationHelpHint)
}

func verifyBridgeCommand(runtime bridgeRuntime) error {
	cmd := runtime.bridgeCommand("--ping")
	cmd.Dir = runtime.workingDir
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		details := strings.TrimSpace(stderr.String())
		if details == "" {
			details = strings.TrimSpace(stdout.String())
		}
		if details != "" {
			return fmt.Errorf("%s bridge command is unavailable. %s\n%s", bridgeRunnerCommand, installationHelpHint, details)
		}
		return fmt.Errorf("%s bridge command is unavailable. %s", bridgeRunnerCommand, installationHelpHint)
	}
	return nil
}

func verifyPythonBridge(runtime bridgeRuntime) error {
	cmd := runtime.bridgeCommand("--ping")
	cmd.Dir = runtime.workingDir
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		details := strings.TrimSpace(stderr.String())
		if details == "" {
			details = strings.TrimSpace(stdout.String())
		}

		if details != "" {
			return fmt.Errorf("python bridge runtime is unavailable. %s\n%s", installationHelpHint, details)
		}
		return fmt.Errorf("python bridge runtime is unavailable. %s", installationHelpHint)
	}

	return nil
}

func findPythonBinary() (string, bool) {
	for _, candidate := range []string{"python3", "python"} {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, true
		}
	}
	return "", false
}

func (runtime bridgeRuntime) pythonCommand(args ...string) *exec.Cmd {
	if runtime.useUv {
		commandArgs := append([]string{"run", "python"}, args...)
		return exec.Command(runtime.uvBinary, commandArgs...)
	}
	return exec.Command(runtime.pythonBinary, args...)
}

func (runtime bridgeRuntime) bridgeCommand(args ...string) *exec.Cmd {
	if runtime.useUv && runtime.useBridgeCmd {
		commandArgs := append([]string{"run", bridgeRunnerCommand}, args...)
		return exec.Command(runtime.uvBinary, commandArgs...)
	}
	commandArgs := append([]string{"-m", bridgeRunnerModule}, args...)
	return runtime.pythonCommand(commandArgs...)
}

func findRuntimeWorkingDir() (string, error) {
	if root, err := common.FindProjectRoot(); err == nil {
		return root, nil
	}
	return os.Getwd()
}

func findProjectRoot() (string, error) {
	return common.FindProjectRoot()
}

func configureHomeContent() (home.Screen, error) {
	if err := home.ConfigureContent(); err != nil {
		return home.Screen{}, err
	}
	return home.NewScreen(home.Actions()), nil
}
