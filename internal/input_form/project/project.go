package project

import (
	"lazyhetzner/internal/input_form"
	"github.com/charmbracelet/bubbles/textinput"
)



func NewProjectForm() input_form.InputForm {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Project name (e.g., production, staging)"
	inputs[0].Focus()
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Hetzner Cloud API token"
	inputs[1].Width = 50
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'

	return input_form.InputForm{
		Inputs:    inputs,
		FocusIdx:  0,
		SubmitBtn: "Add Project",
		CancelBtn: "Cancel",
	}
}
