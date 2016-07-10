package printer

import (
	"html/template"
	"io"
	"os"

	"github.com/fatih/color"
)

var helpers = template.FuncMap{
	"magenta": color.MagentaString,
	"yellow":  color.YellowString,
	"green":   color.GreenString,
	"black":   color.BlackString,
	"white":   color.WhiteString,
	"blue":    color.BlueString,
	"cyan":    color.CyanString,
	"red":     color.RedString,
}

const successTmpl = `
 - {{. | cyan}}
`

const errTmpl = `
 - {{. | red}}
`

// Printer is struct
type Printer struct {
	writer    io.Writer
	errWriter io.Writer
}

// NewPrinter is constructor
func NewPrinter(writer, errWriter io.Writer) *Printer {
	if writer == nil {
		writer = os.Stdout
	}
	if errWriter == nil {
		errWriter = os.Stderr
	}

	return &Printer{
		writer: writer,
	}
}

// Print is print text
func (p *Printer) Print(txt string) {
	t(true).Execute(p.writer, txt)
	return
}

// Error is print error
func (p *Printer) Error(err error) {
	t(false).Execute(p.writer, err.Error())
	return
}

func t(isSuccess bool) *template.Template {
	if isSuccess {
		return template.Must(template.New("item").Funcs(helpers).Parse(successTmpl))
	}
	return template.Must(template.New("item").Funcs(helpers).Parse(errTmpl))
}
