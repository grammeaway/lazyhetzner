package model
// App states
type state int

const (
	StateProjectSelect state = iota
	stateProjectManage
	stateTokenInput
	stateLoading
	stateResourceView
	stateLabelView
	stateContextMenu
	stateError
)
