package cmd
// App states
type state int

const (
	stateProjectSelect state = iota
	stateProjectManage
	stateTokenInput
	stateLoading
	stateResourceView
	stateContextMenu
	stateError
)
