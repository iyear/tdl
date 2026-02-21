package tui

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tclient"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/pkg/consts"
)

type sessionState int

const (
	stateDashboard sessionState = iota
	stateDownloads
	stateConfig
	stateBrowser
	stateBatch
	stateBatchConfirm
	stateLogin
	stateLoginPhone
	stateLoginCode
	stateLoginPassword
	stateExportPrompt
	stateDownloadOptions
	stateDirPicker
	stateAccounts
)

type AccountItem struct {
	Name     string
	IsActive bool
}

func (a AccountItem) Title() string { return a.Name }
func (a AccountItem) Description() string {
	if a.IsActive {
		return "Currently Active"
	}
	return "Press Enter to switch"
}
func (a AccountItem) FilterValue() string { return a.Name }

type DownloadForm struct {
	UrlOrPath string
	IsBatch   bool

	// Fields
	Dir      textinput.Model
	Template textinput.Model

	// bools
	Group    bool
	SkipSame bool
	Takeout  bool
	Desc     bool

	// Advanced
	Threads   textinput.Model
	Limit     textinput.Model
	Pool      textinput.Model
	Delay     textinput.Model
	Reconnect textinput.Model

	// Advanced bools
	Continue bool
	Debug    bool

	ActiveIndex int // 0: Dir, 1: Template, 2-5: Basic Bools, 6-10: Advanced Inputs, 11-12: Adv Bools, 13: Start, 14: Cancel
}

type Model struct {
	state      sessionState
	ActiveTab  int   // 0: Dashboard, 1: Browser, 2: Downloads, 3: Forwarding
	TabHistory []int // Navigation stack for Esc key

	// Browser State
	Dialogs        list.Model
	Messages       list.Model
	Browsing       bool // True if focused heavily on browser
	Pane           int  // 0: Dialogs (Left), 1: Messages (Right)
	SelectedApp    *tclient.App
	LoadingDialogs bool
	IsPaginating   bool
	NextOffsetPeer tg.InputPeerClass
	NextOffsetDate int
	NextOffsetID   int

	LoadingHistory bool
	LoadingExport  bool
	Searching      bool // Global Search input mode

	// Forwarding
	PickingDest   bool // If true, selecting a dialog = forward destination
	ForwardSource []string

	// Export
	ExportInput  textinput.Model
	ExportTarget DialogItem

	// Download Options
	DLForm DownloadForm

	// UI State
	ShowHelp bool

	width    int
	height   int
	quitting bool

	// Components
	spinner  spinner.Model
	list     list.Model
	viewport viewport.Model
	input    textinput.Model

	// Config Editor
	ConfigInputs     []textinput.Model
	ConfigFocusIndex int

	// Batch Processing
	FilePicker filepicker.Model
	BatchPath  string

	// Data
	Namespace     string
	Connected     bool
	BuildInfo     string
	User          *tg.User
	Downloads     map[string]*DownloadItem
	Forwards      map[int64]*ForwardItem
	DownloadList  list.Model // New list for downloads
	ForwardList   list.Model
	StatusMessage string

	// Account Management
	Accounts     []string
	AccountsList list.Model
	kvStorage    kv.Storage

	// Login State
	AuthPhone    textinput.Model
	AuthCode     textinput.Model
	AuthPassword textinput.Model
	AuthCodeHash string

	// System Metrics
	sysCpu float64
	sysMem float64

	// Internal
	storage    storage.Storage
	tuiProgram *tea.Program

	// Persistent Client
	Client       *telegram.Client
	ClientCtx    context.Context
	ClientCancel context.CancelFunc
	clientMu     sync.Mutex
}

type loginMsg struct {
	User *tg.User
	Err  error
}

type ExportMsg struct {
	Path string
	Err  error
}

type ExportProgressMsg int64

type AccountsMsg struct {
	Accounts []string
	Err      error
}

type AccountSwitchedMsg struct {
	Namespace string
	Storage   storage.Storage
	Err       error
}

type sysTickMsg time.Time

func sysTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return sysTickMsg(t)
	})
}

func fetchSysMetrics() (float64, float64) {
	var c float64
	var m float64

	p, err := cpu.Percent(0, false)
	if err == nil && len(p) > 0 {
		c = p[0]
	}

	vm, err := mem.VirtualMemory()
	if err == nil {
		m = vm.UsedPercent
	}
	return c, m
}

func NewModel(root kv.Storage, s storage.Storage, ns string) *Model {
	// Initialize Theme
	themeName := viper.GetString("theme.name")
	if themeName == "" {
		themeName = "Default"
	}
	ApplyTheme(themeName)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	ti := textinput.New()
	ti.Placeholder = "Enter Telegram Link..."
	ti.CharLimit = 156
	ti.Width = 40

	// Initialize Browser Lists
	dList := list.New([]list.Item{}, ItemDelegate{}, 0, 0)
	dList.Title = "Chats"
	dList.SetShowHelp(false)

	mList := list.New([]list.Item{}, ItemDelegate{}, 0, 0)
	mList.Title = "Messages"
	mList.SetShowHelp(false)

	// File Picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.CurrentDirectory, _ = os.Getwd()
	fp.Height = 10

	// Export Input
	ei := textinput.New()
	ei.Placeholder = "Enter filename (e.g. data.json)"
	ei.CharLimit = 50
	ei.Width = 40

	// Download List
	dlList := list.New([]list.Item{}, ItemDelegate{}, 0, 0)
	dlList.Title = "Downloads"
	dlList.SetShowHelp(false)

	// Download Form Inputs
	dirInput := textinput.New()
	dirInput.Placeholder = "downloads"
	dirInput.CharLimit = 100
	dirInput.Width = 40
	// Pre-fill from config
	defaultDir := viper.GetString("download_dir")
	if defaultDir == "" {
		defaultDir = "downloads"
	}
	dirInput.SetValue(defaultDir)

	tmplInput := textinput.New()
	tmplInput.Placeholder = "{{ .Index }}-{{ .ID }}"
	tmplInput.CharLimit = 100
	tmplInput.Width = 40
	// Pre-fill
	defaultTmpl := viper.GetString(consts.FlagDlTemplate)
	if defaultTmpl == "" {
		defaultTmpl = "{{ .Index }}-{{ .ID }}"
	}
	tmplInput.SetValue(defaultTmpl)

	// Advanced Inputs
	// Helper to create small input
	newAdvInput := func(val string, width int) textinput.Model {
		t := textinput.New()
		t.SetValue(val)
		t.Width = width
		return t
	}

	threads := newAdvInput(strconv.Itoa(viper.GetInt(consts.FlagThreads)), 5)
	limit := newAdvInput(strconv.Itoa(viper.GetInt(consts.FlagLimit)), 5)
	pool := newAdvInput(strconv.Itoa(viper.GetInt(consts.FlagPoolSize)), 5)
	delay := newAdvInput(viper.GetDuration(consts.FlagDelay).String(), 10)
	reconnect := newAdvInput(viper.GetDuration(consts.FlagReconnectTimeout).String(), 10)

	// Auth Inputs
	authPhone := textinput.New()
	authPhone.Placeholder = "Phone number (e.g., +1234567890)"
	authPhone.Width = 30

	authCode := textinput.New()
	authCode.Placeholder = "Verification code"
	authCode.Width = 20

	authPassword := textinput.New()
	authPassword.Placeholder = "2FA Password (if any)"
	authPassword.EchoMode = textinput.EchoPassword
	authPassword.EchoCharacter = '*'
	authPassword.Width = 30

	accList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	accList.Title = "Authentication Sessions"
	accList.SetShowHelp(false)

	fwList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	fwList.Title = "Active Forwarding Clones"
	fwList.SetShowHelp(false)

	return &Model{
		state:        stateDashboard,
		ActiveTab:    0, // Dashboard default
		Dialogs:      dList,
		Messages:     mList,
		Pane:         0, // Start with Dialogs focused
		spinner:      sp,
		Namespace:    ns,
		BuildInfo:    consts.Version,
		storage:      s,
		kvStorage:    root,
		Downloads:    make(map[string]*DownloadItem),
		Forwards:     make(map[int64]*ForwardItem),
		DownloadList: dlList,
		ForwardList:  fwList,
		AccountsList: accList,
		input:        ti,
		ExportInput:  ei,
		FilePicker:   fp,
		AuthPhone:    authPhone,
		AuthCode:     authCode,
		AuthPassword: authPassword,
		DLForm: DownloadForm{
			Dir:       dirInput,
			Template:  tmplInput,
			Group:     viper.GetBool("group"),
			SkipSame:  viper.GetBool("skip_same"),
			Takeout:   viper.GetBool("takeout"),
			Desc:      viper.GetBool("desc"),
			Threads:   threads,
			Limit:     limit,
			Pool:      pool,
			Delay:     delay,
			Reconnect: reconnect,
			Continue:  viper.GetBool("continue"),
			Debug:     viper.GetBool("debug"),
		},
	}
}

func (m *Model) SetProgram(p *tea.Program) {
	m.tuiProgram = p
}

func (m *Model) Init() tea.Cmd {
	// Initialize Status
	m.StatusMessage = "Connecting to Telegram..."

	return tea.Batch(
		m.spinner.Tick,
		m.startClient, // Start the persistent connection
		m.GetAccounts(),
		m.GetDialogs(nil, 0, 0), // Initial load
		sysTick(),
	)
}

func (m *Model) startClient() tea.Msg {
	m.clientMu.Lock()
	defer m.clientMu.Unlock()

	// Cleanup existing client if any
	if m.ClientCancel != nil {
		m.ClientCancel()
	}

	// Create context for the client lifecycle
	m.ClientCtx, m.ClientCancel = context.WithCancel(context.Background())

	logToFile("StartClient: Context created")

	// Create the client instance
	opts := tclient.Options{
		KV:               m.storage,
		Proxy:            viper.GetString(consts.FlagProxy),
		NTP:              viper.GetString(consts.FlagNTP),
		ReconnectTimeout: viper.GetDuration(consts.FlagReconnectTimeout),
	}

	logToFile(fmt.Sprintf("StartClient: Options - Proxy: %v, NTP: %v", opts.Proxy != "", opts.NTP))

	var err error
	m.Client, err = tclient.New(m.ClientCtx, opts, false)
	if err != nil {
		return loginMsg{Err: err}
	}

	// Run the client in a goroutine
	// We use a channel to signal when the client is ready/authorized effectively?
	// Actually client.Run blocks. We need to run it and then perform a self check.
	// But checkLogin expects a return based on 'User'.

	// Strategy:
	// 1. Start client.Run in goroutine.
	// 2. Wait for it to be ready (Auth Status).
	// 3. Fetch Self.
	// 4. Return loginMsg.

	readyCh := make(chan struct{})
	errCh := make(chan error)

	go func() {
		logToFile("ClientGoroutine: Starting Run")
		err := m.Client.Run(m.ClientCtx, func(ctx context.Context) error {
			// Signal ready
			logToFile("ClientGoroutine: Connected! Signaling Ready")
			close(readyCh)
			// Choose to block until context is done
			<-ctx.Done()
			logToFile("ClientGoroutine: Context Done, Exiting")
			return ctx.Err()
		})
		if err != nil && err != context.Canceled {
			logToFile("ClientGoroutine: Error: " + err.Error())
			errCh <- err
		} else {
			logToFile("ClientGoroutine: Stopped Gracefully")
		}
	}()

	// Wait for ready or error
	logToFile("StartClient: Waiting for Ready signal...")
	select {
	case <-readyCh:
		logToFile("StartClient: Received Ready signal")
		// Client is running. Now check auth.
		// We need to use m.ClientCtx or a sub-context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		status, err := m.Client.Auth().Status(ctx)
		if err != nil {
			return loginMsg{Err: err}
		}

		if !status.Authorized {
			return loginMsg{Err: fmt.Errorf("not authorized")}
		}

		// Fetch Self
		self, err := m.Client.Self(ctx)
		if err != nil {
			return loginMsg{Err: err}
		}
		return loginMsg{User: self}

	case err := <-errCh:
		return loginMsg{Err: err}
	case <-time.After(15 * time.Second):
		return loginMsg{Err: fmt.Errorf("connection timeout")}
	}
}
