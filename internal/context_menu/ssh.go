package context_menu

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"runtime"
	tea "github.com/charmbracelet/bubbletea"
 	"lazyhetzner/internal/message"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// SessionType represents the type of terminal multiplexer session
type SessionType int

const (
	SessionNone SessionType = iota
	SessionTmux
	SessionZellij
)


// SSH action messages
type sshLaunchedMsg struct{}

type tmuxSSHLaunchedMsg struct{}

type zellijSSHLaunchedMsg struct{}

// SessionInfo holds information about the current session
type SessionInfo struct {
	Type        SessionType
	SessionName string
	WindowName  string
	PaneName    string
}

// detectSession detects if we're running inside tmux or zellij
func detectSession() SessionInfo {
	info := SessionInfo{Type: SessionNone}

	// Check for tmux
	if tmuxSession := os.Getenv("TMUX"); tmuxSession != "" {
		info.Type = SessionTmux
		info.SessionName = os.Getenv("TMUX_SESSION")
		info.WindowName = os.Getenv("TMUX_WINDOW")
		info.PaneName = os.Getenv("TMUX_PANE")

		// If session name is empty, try to get it via tmux command
		if info.SessionName == "" {
			if cmd := exec.Command("tmux", "display-message", "-p", "#S"); cmd != nil {
				if output, err := cmd.Output(); err == nil {
					info.SessionName = strings.TrimSpace(string(output))
				}
			}
		}
		return info
	}

	// Check for zellij
	if zellijSession := os.Getenv("ZELLIJ"); zellijSession != "" {
		info.Type = SessionZellij
		info.SessionName = os.Getenv("ZELLIJ_SESSION_NAME")
		return info
	}

	return info
}

// getSSHMenuItems returns context menu items based on the current session
func getSSHMenuItems(sessionInfo SessionInfo) []contextMenuItem {
	baseItems := []contextMenuItem{
		{label: "📋 Copy Public IP", action: "copy_public_ip"},
		{label: "📋 Copy Private IP", action: "copy_private_ip"},
	}

	switch sessionInfo.Type {
	case SessionTmux:
		return append(baseItems, []contextMenuItem{
			{label: "🪟 SSH (New tmux window)", action: "ssh_tmux_window"},
			{label: "📱 SSH (New tmux pane)", action: "ssh_tmux_pane"},
			{label: "🔗 SSH (New terminal)", action: "ssh_new_terminal"},
			{label: "🔗 SSH (Current terminal)", action: "ssh_current_terminal"},
		}...)

	case SessionZellij:
		return append(baseItems, []contextMenuItem{
			{label: "🪟 SSH (New zellij tab)", action: "ssh_zellij_tab"},
			{label: "📱 SSH (New zellij pane)", action: "ssh_zellij_pane"},
			{label: "🔗 SSH (New terminal)", action: "ssh_new_terminal"},
			{label: "🔗 SSH (Current terminal)", action: "ssh_current_terminal"},
		}...)

	default:
		return append(baseItems, []contextMenuItem{
			{label: "🔗 SSH (New terminal)", action: "ssh_new_terminal"},
			{label: "🔗 SSH (Current terminal)", action: "ssh_current_terminal"},
		}...)
	}
}


func launchSSH(ip string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd

		// Determine terminal based on OS
		switch runtime.GOOS {
		case "darwin": // macOS
			cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "ssh root@%s"`, ip))
		case "linux":
			// Try common terminal emulators
			terminals := []string{"gnome-terminal", "konsole", "xterm", "alacritty", "kitty", "foot"}
			for _, term := range terminals {
				if _, err := exec.LookPath(term); err == nil {
					switch term {
					case "gnome-terminal":
						cmd = exec.Command(term, "--", "ssh", fmt.Sprintf("root@%s", ip))
					case "konsole":
						cmd = exec.Command(term, "-e", "ssh", fmt.Sprintf("root@%s", ip))
					default:
						cmd = exec.Command(term, "-e", "ssh", fmt.Sprintf("root@%s", ip))
					}
					break
				}
			}
		case "windows":
			// Use Windows Terminal if available, fallback to Powershell, otherwise cmd
			if _, err := exec.LookPath("wt"); err == nil {
				cmd = exec.Command("wt", "ssh", fmt.Sprintf("root@%s", ip))
			} else if _, err := exec.LookPath("powershell"); err == nil {
				cmd = exec.Command("powershell", "-Command", fmt.Sprintf("ssh root@%s", ip))
			} else {
				cmd = exec.Command("cmd", "/C", fmt.Sprintf("ssh root@%s", ip))
			}
		}

		if cmd == nil {
			return message.ErrorMsg{fmt.Errorf("no suitable terminal found")}
		}

		err := cmd.Start()
		if err != nil {
			return message.ErrorMsg{err}
		}

		return sshLaunchedMsg{}
	}
}


// launchSSHInTmuxWindow launches SSH in a new tmux window
func launchSSHInTmuxWindow(ip string) tea.Cmd {
	return func() tea.Msg {
		windowName := fmt.Sprintf("ssh-%s", strings.ReplaceAll(ip, ".", "-"))
		cmd := exec.Command("tmux", "new-window", "-n", windowName, fmt.Sprintf("ssh root@%s", ip))

		if err := cmd.Run(); err != nil {
			return message.ErrorMsg{fmt.Errorf("failed to create tmux window: %w", err)}
		}

		return message.StatusMsg(fmt.Sprintf("🪟 SSH session launched in new tmux window: %s", windowName))
	}
}

// launchSSHInTmuxPane launches SSH in a new tmux pane
func launchSSHInTmuxPane(ip string) tea.Cmd {
	return func() tea.Msg {
		// Split the current pane horizontally and run SSH
		cmd := exec.Command("tmux", "split-window", "-h", fmt.Sprintf("ssh root@%s", ip))

		if err := cmd.Run(); err != nil {
			return message.ErrorMsg{fmt.Errorf("failed to create tmux pane: %w", err)}
		}

		return message.StatusMsg("📱 SSH session launched in new tmux pane")
	}
}

// launchSSHInZellijTab launches SSH in a new zellij tab
func launchSSHInZellijTab(ip string) tea.Cmd {
	return func() tea.Msg {
		tabName := fmt.Sprintf("ssh-%s", strings.ReplaceAll(ip, ".", "-"))

		// Create new tab with SSH command
		cmd := exec.Command("zellij", "action", "new-tab", "--name", tabName, "--", "ssh", fmt.Sprintf("root@%s", ip))

		if err := cmd.Run(); err != nil {
			return message.ErrorMsg{fmt.Errorf("failed to create zellij tab: %w", err)}
		}

		return message.StatusMsg(fmt.Sprintf("🪟 SSH session launched in new zellij tab: %s", tabName))
	}
}

// launchSSHInZellijPane launches SSH in a new zellij pane
func launchSSHInZellijPane(ip string) tea.Cmd {
	return func() tea.Msg {
		// Split the current pane and run SSH
		cmd := exec.Command("zellij", "action", "new-pane", "--", "ssh", fmt.Sprintf("root@%s", ip))

		if err := cmd.Run(); err != nil {
			return message.ErrorMsg{fmt.Errorf("failed to create zellij pane: %w", err)}
		}

		return message.StatusMsg("📱 SSH session launched in new zellij pane")
	}
}

func launchSSHInSameTerminal(ip string) tea.Cmd {

	return tea.ExecProcess(exec.Command("ssh", fmt.Sprintf("root@%s", ip)), func(err error) tea.Msg {
		if err != nil {
			return message.ErrorMsg{err}
		}
		return sshLaunchedMsg{}
	})
}


func (m *model) initSessionInfo() {
	m.sessionInfo = detectSession()
}

func createServerContextMenu(server *hcloud.Server, sessionInfo SessionInfo) contextMenu {
	return contextMenu{
		items:        getSSHMenuItems(sessionInfo),
		selectedItem: 0,
		server:       server,
	}
}

func handleSSHAction(action string, server *hcloud.Server, sessionInfo SessionInfo) tea.Cmd {
	if server.PublicNet.IPv4.IP == nil {
		return func() tea.Msg {
			return message.ErrorMsg{fmt.Errorf("server has no public IP")}
		}
	}

	ip := server.PublicNet.IPv4.IP.String()

	switch action {
	case "ssh_tmux_window":
		return launchSSHInTmuxWindow(ip)
	case "ssh_tmux_pane":
		return launchSSHInTmuxPane(ip)
	case "ssh_zellij_tab":
		return launchSSHInZellijTab(ip)
	case "ssh_zellij_pane":
		return launchSSHInZellijPane(ip)
	case "ssh_new_terminal":
		return launchSSH(ip)
	case "ssh_current_terminal":
		return launchSSHInSameTerminal(ip)
	default:
		return nil
	}
}
