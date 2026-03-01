package tui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/app/forward"
	"github.com/iyear/tdl/core/forwarder"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/viper"
)

func (m *Model) startDownload(url string) tea.Cmd {
	storage := m.storage // Capture for thread safety
	prog := m.tuiProgram
	return func() tea.Msg {
		if url == "" {
			return nil
		}

		// In a real app we'd want to manage context cancellation
		ctx, cancel := context.WithCancel(context.Background())

		// Prepare Options from Form
		dir := m.DLForm.Dir.Value()
		if dir == "" {
			dir = viper.GetString("download_dir")
			if dir == "" {
				dir = "downloads"
			}
		}

		tmpl := m.DLForm.Template.Value()
		if tmpl == "" {
			tmpl = viper.GetString(consts.FlagDlTemplate)
		}

		// Parse Advanced Options
		threads, _ := strconv.Atoi(m.DLForm.Threads.Value())
		limit, _ := strconv.Atoi(m.DLForm.Limit.Value())
		poolSize, _ := strconv.Atoi(m.DLForm.Pool.Value())
		delay, _ := time.ParseDuration(m.DLForm.Delay.Value())
		reconnect, _ := time.ParseDuration(m.DLForm.Reconnect.Value())

		opts := dl.Options{
			URLs:             []string{url},
			Dir:              dir,
			Template:         tmpl,
			Group:            m.DLForm.Group,
			SkipSame:         m.DLForm.SkipSame || m.DLForm.Continue, // Auto-skip if Continuing
			Takeout:          m.DLForm.Takeout,
			Desc:             m.DLForm.Desc,
			Continue:         m.DLForm.Continue,
			Debug:            m.DLForm.Debug,
			Threads:          threads,
			Limit:            limit,
			PoolSize:         poolSize,
			Delay:            delay,
			ReconnectTimeout: reconnect,
		}

		// We need to run this in a way that respects the existing architecture
		// The key challenge is that dl.Run takes existing Client and KV
		// We have KV, but Client is usually created inside tRun or passed in.
		// In our login check we created a client briefly.
		// We should probably keep a persistent client or recreate it.
		// Recreating it is safer for now.

		// Use persistent client
		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			cancel()
			return ProgressMsg{Name: url, Err: fmt.Errorf("client not connected"), IsFinished: true}
		}

		// Inject TUI progress and enable Silent mode
		opts.Silent = true
		opts.ExternalProgress = NewTUIProgress(prog)

		// Send initial message with cancel func
		prog.Send(ProgressMsg{Name: url, Cancel: cancel})

		// Run download
		// dl.Run expects client to be valid.
		// Note: dl.Run might block. We are in a tea.Cmd (goroutine), so it's fine.
		err := dl.Run(logctx.Named(ctx, "dl"), client, storage, opts)
		return ProgressMsg{Name: url, Err: err, IsFinished: true}
	}
}

func (m *Model) startBatchDownload(path string) tea.Cmd {
	storage := m.storage
	prog := m.tuiProgram
	return func() tea.Msg {
		if path == "" {
			return nil
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Prepare Options from Form
		dir := m.DLForm.Dir.Value()
		if dir == "" {
			dir = viper.GetString("download_dir")
			if dir == "" {
				dir = "downloads"
			}
		}

		tmpl := m.DLForm.Template.Value()
		if tmpl == "" {
			tmpl = viper.GetString(consts.FlagDlTemplate)
		}

		// Parse Advanced Options
		threads, _ := strconv.Atoi(m.DLForm.Threads.Value())
		limit, _ := strconv.Atoi(m.DLForm.Limit.Value())
		poolSize, _ := strconv.Atoi(m.DLForm.Pool.Value())
		delay, _ := time.ParseDuration(m.DLForm.Delay.Value())
		reconnect, _ := time.ParseDuration(m.DLForm.Reconnect.Value())

		opts := dl.Options{
			Files:            []string{path},
			Dir:              dir,
			Template:         tmpl,
			Group:            m.DLForm.Group,
			SkipSame:         m.DLForm.SkipSame || m.DLForm.Continue, // Auto-skip if Continuing
			Takeout:          m.DLForm.Takeout,
			Desc:             m.DLForm.Desc,
			Continue:         m.DLForm.Continue,
			Debug:            m.DLForm.Debug,
			Threads:          threads,
			Limit:            limit,
			PoolSize:         poolSize,
			Delay:            delay,
			ReconnectTimeout: reconnect,
		}

		// Use persistent client
		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			cancel()
			return ProgressMsg{Name: path, Err: fmt.Errorf("client not connected"), IsFinished: true}
		}

		opts.Silent = true
		opts.ExternalProgress = NewTUIProgress(prog)

		// Send initial message with cancel func
		prog.Send(ProgressMsg{Name: path, Cancel: cancel})

		err := dl.Run(logctx.Named(ctx, "dl"), client, storage, opts)
		return ProgressMsg{Name: path, Err: err, IsFinished: true}
	}
}

func (m *Model) startExport(d DialogItem, filename string) tea.Cmd {
	storage := m.storage
	return func() tea.Msg {
		ctx := context.Background()

		// Use provided filename
		if filename == "" {
			filename = fmt.Sprintf("%d.json", d.PeerID)
		}

		// Setup Options
		opts := chat.ExportOptions{
			Type:   chat.ExportTypeTime,
			Input:  []int{0, math.MaxInt}, // All history
			Output: filename,
			Chat:   strconv.FormatInt(d.PeerID, 10),
			Silent: true,
			Filter: "true",
			Progress: func(count int64) {
				if m.tuiProgram != nil {
					m.tuiProgram.Send(ExportProgressMsg(count))
				}
			},
		}

		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			return ExportMsg{Err: fmt.Errorf("client not connected")}
		}

		err := chat.Export(logctx.Named(ctx, "export"), client, storage, opts)

		return ExportMsg{Path: filename, Err: err}
	}
}

func (m *Model) GetAccounts() tea.Cmd {
	return func() tea.Msg {
		if m.kvStorage == nil {
			return nil
		}
		items, err := m.kvStorage.Namespaces()
		if err != nil {
			return AccountsMsg{Err: err}
		}
		return AccountsMsg{Accounts: items}
	}
}

func (m *Model) SwitchAccount(ns string) tea.Cmd {
	return func() tea.Msg {
		if m.kvStorage == nil {
			return nil
		}
		kvd, err := m.kvStorage.Open(ns)
		return AccountSwitchedMsg{Namespace: ns, Storage: kvd, Err: err}
	}
}

func (m *Model) startForward(dest string, sources []string) tea.Cmd {
	storage := m.storage
	return func() tea.Msg {
		ctx := context.Background()

		opts := forward.Options{
			From:             sources,
			To:               dest,                 // Destination is now dynamic
			Mode:             forwarder.ModeDirect, // Switch to instant native server-side forwarding
			Silent:           true,
			ExternalProgress: NewTUIForwardProgress(m.tuiProgram),
		}

		// Parse Thread Option from active Form for fast clone fallback
		threads, errParse := strconv.Atoi(m.DLForm.Threads.Value())
		if errParse != nil || threads < 1 {
			threads = 4 // Default to CLI's concurrent speed
		}
		viper.Set(consts.FlagThreads, threads)

		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			return ExportMsg{Err: fmt.Errorf("client not connected")}
		}

		err := forward.Run(logctx.Named(ctx, "forward"), client, storage, opts)

		return ExportMsg{Path: "Forwarded", Err: err} // Reusing ExportMsg for simplicity for now
	}
}

func (m *Model) SearchPeers(query string) tea.Cmd {
	return func() tea.Msg {
		if query == "" {
			return nil
		}
		ctx := context.Background()

		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			return dialogsMsg{Err: fmt.Errorf("client not connected")}
		}

		var results []DialogItem

		// Use m.Client.API() equivalent but context bound
		// client.API() returns *tg.Client
		// We can just use tg.NewClient(client)
		raw := tg.NewClient(client)

		res, err := raw.ContactsSearch(ctx, &tg.ContactsSearchRequest{
			Q:     query,
			Limit: 20,
		})
		if err != nil {
			return dialogsMsg{Err: err}
		}

		// Process results
		found := res

		// Helper to find title and input peer
		getTitle := func(peerC tg.PeerClass) (string, int64, tg.InputPeerClass) {
			var id int64
			var title string
			var inputPeer tg.InputPeerClass

			switch p := peerC.(type) {
			case *tg.PeerUser:
				id = p.UserID
				for _, u := range found.Users {
					switch user := u.(type) {
					case *tg.User:
						if user.ID == id {
							title = user.FirstName + " " + user.LastName
							if user.Username != "" {
								title += " (@" + user.Username + ")"
							}
							inputPeer = &tg.InputPeerUser{UserID: id, AccessHash: user.AccessHash}
						}
					}
					if inputPeer != nil {
						break
					}
				}
				if inputPeer == nil {
					inputPeer = &tg.InputPeerUser{UserID: id}
				} // Fallback

			case *tg.PeerChat:
				id = p.ChatID
				for _, c := range found.Chats {
					switch chat := c.(type) {
					case *tg.Chat:
						if chat.ID == id {
							title = chat.Title
						}
					}
					// Chat usually doesn't need access hash for InputPeerChat
					if title != "" {
						break
					}
				}
				inputPeer = &tg.InputPeerChat{ChatID: id}

			case *tg.PeerChannel:
				id = p.ChannelID
				for _, c := range found.Chats {
					switch chat := c.(type) {
					case *tg.Channel:
						if chat.ID == id {
							title = chat.Title
							inputPeer = &tg.InputPeerChannel{ChannelID: id, AccessHash: chat.AccessHash}
						}
					}
					if inputPeer != nil {
						break
					}
				}
				if inputPeer == nil {
					inputPeer = &tg.InputPeerChannel{ChannelID: id}
				}
			}

			if title == "" {
				title = fmt.Sprintf("Unknown#%d", id)
			}
			return title, id, inputPeer
		}

		// Process Results
		results = make([]DialogItem, 0, len(found.Results))
		for _, p := range found.Results {
			title, id, inputPeer := getTitle(p)
			results = append(results, DialogItem{
				Title:  title,
				PeerID: id,
				Peer:   inputPeer,
			})
		}
		return dialogsMsg{Dialogs: results}
	}
}

// Authentication Flows

type authMsg struct {
	State sessionState
	Hash  string
	Err   error
}

func (m *Model) loginSendCode(phone string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			return authMsg{Err: fmt.Errorf("client not connected")}
		}

		// Initialize Auth Flow
		res, err := client.Auth().SendCode(ctx, phone, auth.SendCodeOptions{})
		if err != nil {
			return authMsg{Err: err}
		}

		var hash string
		if sentCode, ok := res.(*tg.AuthSentCode); ok {
			hash = sentCode.PhoneCodeHash
		}

		return authMsg{State: stateLoginCode, Hash: hash}
	}
}

func (m *Model) loginVerifyCode(code string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			return authMsg{Err: fmt.Errorf("client not connected")}
		}

		// AuthCodeHash will be read from Model in update.go and passed down or stored in Model.
		// Wait, I need m.AuthCodeHash.
		_, err := client.Auth().SignIn(ctx, m.AuthPhone.Value(), code, m.AuthCodeHash)
		if err != nil {
			// Check if password is required
			if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
				return authMsg{State: stateLoginPassword}
			}
			return authMsg{Err: err}
		}

		return authMsg{State: stateDashboard} // Login successful
	}
}

func (m *Model) loginVerifyPassword(password string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		m.clientMu.Lock()
		client := m.Client
		m.clientMu.Unlock()

		if client == nil {
			return authMsg{Err: fmt.Errorf("client not connected")}
		}

		_, err := client.Auth().Password(ctx, password)
		if err != nil {
			return authMsg{Err: err}
		}

		return authMsg{State: stateDashboard} // Login successful
	}
}
