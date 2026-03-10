package model

// App states
type state int

const (
	StateProjectSelect state = iota
	stateProjectManage
	stateTerminalConfig
	stateTokenInput
	stateLoading
	stateResourceView
	stateLabelView
	stateContextMenu
	stateLoadBalancerServiceView
	stateLoadBalancerTargetView
	stateFirewallRuleView
	stateNetworkSubnetView
	stateServerDetailView
	stateError
)
