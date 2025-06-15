package server

import (
	"fmt"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hetznercloud/hcloud-go/hcloud"
	ctm "lazyhetzner/internal/context_menu"
	"lazyhetzner/internal/message"
	"lazyhetzner/internal/resource"
	"lazyhetzner/internal/resource/label"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SessionType represents the type of terminal multiplexer session
type SessionType int

const (
	SessionNone SessionType = iota
	SessionTmux
	SessionZellij
)

// SSH Action messages
type SshLaunchedMsg struct{}

type TmuxSSHLaunchedMsg struct{}

type ZellijSSHLaunchedMsg struct{}

// SessionInfo holds information about the current session
type SessionInfo struct {
	Type        SessionType
	SessionName string
	WindowName  string
	PaneName    string
}

// Detects if running inside tmux or zellij
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

// Returns context menu Items for servers
func getServerMenuItems(sessionInfo SessionInfo) []ctm.ContextMenuItem {
	baseItems := []ctm.ContextMenuItem{
		// add action for canceling (i.e., closing) the context menu
		{Label: "❌ Cancel", Action: "cancel"},
		{Label: "🔖 View Labels", Action: "view_labels"},
		{Label: "📋 Copy Public IP", Action: "copy_public_ip"},
		{Label: "📋 Copy Private IP", Action: "copy_private_ip"},
	}

	switch sessionInfo.Type {
	case SessionTmux:
		return append(baseItems, []ctm.ContextMenuItem{
			{Label: "🪟 SSH (New tmux window)", Action: "ssh_tmux_window"},
			{Label: "📱 SSH (New tmux pane)", Action: "ssh_tmux_pane"},
			{Label: "🔗 SSH (New terminal)", Action: "ssh_new_terminal"},
			{Label: "🔗 SSH (Current terminal)", Action: "ssh_current_terminal"},
		}...)

	case SessionZellij:
		return append(baseItems, []ctm.ContextMenuItem{
			{Label: "🪟 SSH (New zellij tab)", Action: "ssh_zellij_tab"},
			{Label: "📱 SSH (New zellij pane)", Action: "ssh_zellij_pane"},
			{Label: "🔗 SSH (New terminal)", Action: "ssh_new_terminal"},
			{Label: "🔗 SSH (Current terminal)", Action: "ssh_current_terminal"},
		}...)

	default:
		return append(baseItems, []ctm.ContextMenuItem{
			{Label: "🔗 SSH (New terminal)", Action: "ssh_new_terminal"},
			{Label: "🔗 SSH (Current terminal)", Action: "ssh_current_terminal"},
		}...)
	}
}

// launchSSH launches SSH in a new terminal window based on the OS
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

		return SshLaunchedMsg{}
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
		cmd := exec.Command("zellij", "Action", "new-tab", "--name", tabName, "--", "ssh", fmt.Sprintf("root@%s", ip))

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
		cmd := exec.Command("zellij", "Action", "new-pane", "--", "ssh", fmt.Sprintf("root@%s", ip))

		if err := cmd.Run(); err != nil {
			return message.ErrorMsg{fmt.Errorf("failed to create zellij pane: %w", err)}
		}

		return message.StatusMsg("📱 SSH session launched in new zellij pane")
	}
}

// launchSSHInSameTerminal launches SSH in the same terminal
func launchSSHInSameTerminal(ip string) tea.Cmd {

	return tea.ExecProcess(exec.Command("ssh", fmt.Sprintf("root@%s", ip)), func(err error) tea.Msg {
		if err != nil {
			return message.ErrorMsg{err}
		}
		return SshLaunchedMsg{}
	})
}

func getServerLabels(server *hcloud.Server) map[string]string {
	labels := make(map[string]string)
	for key, value := range server.Labels {
		labels[key] = value
	}
	return labels
}

func CreateServerContextMenu(server *hcloud.Server) ctm.ContextMenu {
	sessionInfo := detectSession()
	return ctm.ContextMenu{
		Items:        getServerMenuItems(sessionInfo),
		SelectedItem: 0,
		ResourceType: resource.ResourceServers,
		ResourceID:   server.ID,
	}
}

func handleSSHAction(Action string, server *hcloud.Server, sessionInfo SessionInfo) tea.Cmd {
	if server.PublicNet.IPv4.IP == nil {
		return func() tea.Msg {
			return message.ErrorMsg{fmt.Errorf("server has no public IP")}
		}
	}

	ip := server.PublicNet.IPv4.IP.String()

	switch Action {
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

func ExecuteServerContextAction(selectedAction string, server *hcloud.Server) tea.Cmd {
	sessionInfo := detectSession()

	// Handle SSH actions based on the selected action
	if strings.HasPrefix(selectedAction, "ssh_") {
		return handleSSHAction(selectedAction, server, sessionInfo)
	}

	// Handle other actions here if needed
	switch selectedAction {
	case "cancel":
		return func() tea.Msg {
			return message.CancelCtxMenuMsg{}
		}
	case "view_labels":
		labels := getServerLabels(server)
		if len(labels) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No labels found for this server.")
			}
		}
		return func() tea.Msg {
			return label.LabelsLoadedMsg{
				Labels:              labels,
				RelatedResourceName: fmt.Sprintf("Server: %s", server.Name),
				RelatedResourceType: resource.ResourceServers,
			}
		}
	case "copy_public_ip":
		return func() tea.Msg {
			if server.PublicNet.IPv4.IP == nil {
				return message.ErrorMsg{fmt.Errorf("server has no public IP")}
			}
			if err := clipboard.WriteAll(server.PublicNet.IPv4.IP.String()); err != nil {
				return message.ErrorMsg{err}
			}
			return message.ClipboardCopiedMsg(server.PublicNet.IPv4.IP.String())
		}
	case "copy_private_ip":
		if server.PrivateNet[0].IP != nil {
			return func() tea.Msg {
				if err := clipboard.WriteAll(server.PrivateNet[0].IP.String()); err != nil {
					return message.ErrorMsg{err}
				}

				return message.ClipboardCopiedMsg(server.PrivateNet[0].IP.String())
			}
		}
		return func() tea.Msg {
			return message.ErrorMsg{fmt.Errorf("server has no private IP")}
		}
	default:
		return nil
	}
}
