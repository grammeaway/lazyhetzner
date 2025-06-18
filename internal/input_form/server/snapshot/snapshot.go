package snapshot 



import (
	"lazyhetzner/internal/input_form"
	"github.com/charmbracelet/bubbles/textinput"
)



func NewServerSnapshot() input_form.InputForm {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Snashot name (e.g., my-server-snapshot)"
	inputs[0].Focus()
	inputs[0].Width = 40

	return input_form.InputForm{
		Inputs:    inputs,
		FocusIdx:  0,
		SubmitBtn: "Create Snapshot",
		CancelBtn: "Cancel",
	}
}
