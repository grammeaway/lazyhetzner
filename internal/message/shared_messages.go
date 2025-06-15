package message

// Shared common messages for the application
type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string { return e.Err.Error() }

type StatusMsg string

type ClipboardCopiedMsg string


type CancelCtxMenuMsg struct{}

