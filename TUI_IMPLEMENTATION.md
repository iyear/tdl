# TDL TUI Implementation Documentation

This document serves as a comprehensive guide to the Terminal User Interface (TUI) implementation for `tdl`, built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

**Branch**: `feature/tui` (This branch contains all TUI-related changes).

---

## 1. Architecture Overview
The TUI follows the **Elm Architecture** pattern (Model-View-Update), implemented via the Bubble Tea framework.

### Components
*   **Model (`app/tui/model.go`)**: Stores the application state.
    *   `state`: Current view (Dashboard, Downloads, etc.).
    *   `Downloads`: Map of active downloads tracked by the UI.
    *   `input`: Text input component for URLs.
    *   `tuiProgram`: Reference to the running Tea program (essential for sending async messages).
    *   `storage`: Connection to the local KV store (BoltDB).

*   **View (`app/tui/view.go`)**: Renders the UI as a string.
    *   Uses `lipgloss` for styling (borders, colors).
    *   Conditional rendering based on `m.state`.

*   **Update (`app/tui/update.go`)**: Handles messages and events.
    *   `tea.KeyMsg`: Keyboard input.
    *   `ProgressMsg`: Custom message sent from the core downloader to update UI bars.

*   **Controller (`app/tui/controller.go`)**: Bridges the TUI with `tdl` core logic.
    *   `startDownload(url string)`: Initiates a download by calling `tclient.New` and `dl.Run`.

---

## 2. Core Integration (Modifications to TDL)
To enable the TUI without rewriting the entire downloader, we introduced non-intrusive hooks into the core `dl` package.

### `app/dl/dl.go`
*   **Added `Silent` flag**: When true, suppresses standard console output (logs, progress bars) to prevent TUI glitches.
*   **Added `ExternalProgress`**: An optional `downloader.Progress` interface field in `Options`. The TUI injects its own listener here.

### `app/dl/progress.go`
*   **Hooked `OnAdd`, `OnDownload`, `OnDone`**: If `ExternalProgress` is present, these methods forward the event to it *before* handling standard console output.
*   **Safety**: Allows the TUI to receive granular updates while keeping the core logic intact.

---

## 3. TUI Progress Handling (`app/tui/progress.go`)
*   Implements `downloader.Progress` interface.
*   Converts `downloader.Elem` (which wraps `*iterElem`) into `ProgressMsg`.
*   **Crucial Fix**: Uses type assertion `elem.To().(interface{ Name() string })` to extract the filename from the underlying `os.File`, as `io.WriterAt` does not expose `Name()`.

---

## 4. Build & Dependencies
*   **New Dependencies**:
    *   `github.com/charmbracelet/bubbletea`
    *   `github.com/charmbracelet/bubbles`
    *   `github.com/charmbracelet/lipgloss`
*   **Build Command**: `go build -o tdl.exe .`

---

## 5. Future Roadmap (What's Left)
The foundational work is complete. The following features were planned but not fully implemented or polished:

1.  **Selectable Downloads**: Allow navigating the download list with Up/Down arrows and interacting (Pause/Cancel).
2.  **Rich Dashboard**: Show bandwidth usage graph, storage stats, and detailed user info.
3.  **File Explorer**: Browse local directory for saving downloads.
4.  **Login Flow**: Currently, if not logged in, the TUI asks you to run `tdl login`. A native TUI login form (phone number input -> code input) would be a major UX improvement.
5.  **Configuration**: Settings screen to toggle `Silent` mode, change download directory, etc.

---

## 6. How to Resume Work
If the local folder is deleted, clone the repo and checkout the `feature/tui` branch.

```bash
git clone https://github.com/harshbhardwaj77/tdl.git
cd tdl
git checkout feature/tui
go mod tidy
go build .
```
