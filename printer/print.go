package printer

import (
	"io"
	"log"
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

const successTmpl = `{{.Repo }}
  {{.Msg }}
`

const errTmpl = `{{.Repo }}
  {{.Msg | red}}
`

const cmdTmpl = `Git commad is
  {{.Cmd | green}} {{.Ops | green}}
`

const msgTmpl = `{{.Msg | green}}
`

const msgErrTmpl = `{{.Msg | red}}
`

const repoErrTmpl = `
{{.Msg | red}}
  {{ range .Repos }}- {{ . }}
  {{ end }}`

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
	err := t.Execute(p.writer, cmds{Cmd: cmd, Ops: strings.Join(options, " ")})
	if err != nil {
		log.Println(err)
	}
}

// PrintMsg prints message
func (p *Printer) PrintMsg(msg string) {
	type message struct {
		Msg string
	}
	t := template.Must(template.New("msg").Funcs(helpers).Parse(msgTmpl))
	err := t.Execute(p.writer, message{Msg: msg})
	if err != nil {
		log.Println(err)
	}
}

// PrintMsgErr prints error message
func (p *Printer) PrintMsgErr(msg string) {
	type message struct {
		Msg string
	}
	t := template.Must(template.New("msg").Funcs(helpers).Parse(msgErrTmpl))
	err := t.Execute(p.writer, message{Msg: msg})
	if err != nil {
		log.Println(err)
	}
}

// PrintRepoErr prints error message
func (p *Printer) PrintRepoErr(msg string, repos []string) {
	type message struct {
		Msg   string
		Repos []string
	}
	t := template.Must(template.New("msg").Funcs(helpers).Parse(repoErrTmpl))
	err := t.Execute(p.writer, message{Msg: msg, Repos: repos})
	if err != nil {
		log.Println(err)
	}
}

// Print prints result
func (p *Printer) Print(res Result) {
	err := t(true).Execute(p.writer, res)
	if err != nil {
		log.Println(err)
	}
}

// Error prints error
func (p *Printer) Error(res Result) {
	res.Msg = res.Err.Error()
	err := t(false).Execute(p.errWriter, res)
	if err != nil {
		log.Println(err)
	}
}

func t(isSuccess bool) *template.Template {
	if isSuccess {
		return template.Must(template.New("item").Funcs(helpers).Parse(successTmpl))
	}
	return template.Must(template.New("item").Funcs(helpers).Parse(errTmpl))
}
