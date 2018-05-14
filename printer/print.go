package printer

import (
	"io"
	"os"
	"strings"
	"text/template"

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
{{.Repo }}
  {{.Msg }}
`

const errTmpl = `
{{.Repo }}
  {{.Msg | red}}
`

const cmdTmpl = `
Git commad is
  {{.Cmd | green}} {{.Ops | green}}
`

const msgTmpl = `
{{.Msg | green}}
`

const msgErrTmpl = `
{{.Msg | red}}
`

const repoErrTmpl = `
{{.Msg | red}}
  {{ range .Repos }}- {{ . }}
  {{ end }}
`

// Printer is struct
type Printer struct {
	writer    io.Writer
	errWriter io.Writer
}

// Result is output
type Result struct {
	Repo string
	Msg  string
	Err  error
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
		writer:    writer,
		errWriter: errWriter,
	}
}

// PrintCmd prints command detail
func (p *Printer) PrintCmd(cmd string, options []string) {
	type cmds struct {
		Cmd string
		Ops string
	}
	t := template.Must(template.New("item").Funcs(helpers).Parse(cmdTmpl))
	t.Execute(p.writer, cmds{Cmd: cmd, Ops: strings.Join(options, " ")})
	return
}

// PrintMsg prints message
func (p *Printer) PrintMsg(msg string) {
	type message struct {
		Msg string
	}
	t := template.Must(template.New("msg").Funcs(helpers).Parse(msgTmpl))
	t.Execute(p.writer, message{Msg: msg})
	return
}

// PrintMsgErr prints error message
func (p *Printer) PrintMsgErr(msg string) {
	type message struct {
		Msg string
	}
	t := template.Must(template.New("msg").Funcs(helpers).Parse(msgErrTmpl))
	t.Execute(p.writer, message{Msg: msg})
	return
}

// PrintRepoErr prints error message
func (p *Printer) PrintRepoErr(msg string, repos []string) {
	type message struct {
		Msg   string
		Repos []string
	}
	t := template.Must(template.New("msg").Funcs(helpers).Parse(repoErrTmpl))
	t.Execute(p.writer, message{Msg: msg, Repos: repos})
	return
}

// Print prints result
func (p *Printer) Print(res Result) {
	t(true).Execute(p.writer, res)
	return
}

// Error prints error
func (p *Printer) Error(res Result) {
	res.Msg = res.Err.Error()
	t(false).Execute(p.errWriter, res)
	return
}

func t(isSuccess bool) *template.Template {
	if isSuccess {
		return template.Must(template.New("item").Funcs(helpers).Parse(successTmpl))
	}
	return template.Must(template.New("item").Funcs(helpers).Parse(errTmpl))
}
