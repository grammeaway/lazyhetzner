package loadbalancer



type loadBalancersLoadedMsg struct {
	loadBalancers []*hcloud.LoadBalancer
}




type loadBalancerItem struct {
	lb *hcloud.LoadBalancer
}

func (i loadBalancerItem) FilterValue() string { return i.lb.Name }
func (i loadBalancerItem) Title() string       { return i.lb.Name }
func (i loadBalancerItem) Description() string {
	status := "🟢 Available"
	if i.lb.PublicNet.Enabled {
		return fmt.Sprintf("%s | %s | Targets: %d", status, i.lb.PublicNet.IPv4.IP.String(), len(i.lb.Targets))
	}
	return fmt.Sprintf("%s | Private only | Targets: %d", status, len(i.lb.Targets))
}



func loadLoadBalancers(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		loadBalancers, err := client.LoadBalancer.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return loadBalancersLoadedMsg{loadBalancers: loadBalancers}
	}
}
