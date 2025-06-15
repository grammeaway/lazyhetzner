package server




type serversLoadedMsg struct {
	servers []*hcloud.Server
}



type serverItem struct {
	server *hcloud.Server
}

func (i serverItem) FilterValue() string { return i.server.Name }
func (i serverItem) Title() string       { return i.server.Name }
func (i serverItem) Description() string {
	var statusDisplay string
	if i.server.Status == hcloud.ServerStatusRunning {
		statusDisplay = "🟢 " + string(i.server.Status)
	} else if i.server.Status == hcloud.ServerStatusStarting || i.server.Status == hcloud.ServerStatusInitializing {
		statusDisplay = "🟡 " + string(i.server.Status)
	} else {
		statusDisplay = "🔴 " + string(i.server.Status)
	}
	return fmt.Sprintf("%s | %s | %s", statusDisplay, i.server.ServerType.Name, i.server.PublicNet.IPv4.IP.String())
}



func loadServers(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		servers, err := client.Server.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return serversLoadedMsg{servers: servers}
	}
}
