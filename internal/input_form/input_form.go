package input_form

import ( 
	"github.com/charmbracelet/bubbles/textinput"
)


// Input forms
type InputForm struct {
	Inputs    []textinput.Model
	FocusIdx  int
	SubmitBtn string
	CancelBtn string
}
