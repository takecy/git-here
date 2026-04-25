package printer

import (
	"io"
	"log"
	"os"
	"strings"
	"sync"
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

const cmdTmpl = `Git command is
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

// Templates are parsed once at package init. text/template.Template.Execute is
// safe for concurrent use, so the same parsed template can be shared across
// goroutines without re-parsing on every call.
var (
	successTpl = template.Must(template.New("success").Funcs(helpers).Parse(successTmpl))
	errTpl     = template.Must(template.New("err").Funcs(helpers).Parse(errTmpl))
	cmdTpl     = template.Must(template.New("cmd").Funcs(helpers).Parse(cmdTmpl))
	msgTpl     = template.Must(template.New("msg").Funcs(helpers).Parse(msgTmpl))
	msgErrTpl  = template.Must(template.New("msgErr").Funcs(helpers).Parse(msgErrTmpl))
	repoErrTpl = template.Must(template.New("repoErr").Funcs(helpers).Parse(repoErrTmpl))
)

// Printer is struct
type Printer struct {
	// mu serialises template.Execute calls so that the multiple Write calls
	// emitted per Execute don't interleave across goroutines, even when
	// writer and errWriter end up at the same TTY.
	mu        sync.Mutex
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
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := cmdTpl.Execute(p.writer, cmds{Cmd: cmd, Ops: strings.Join(options, " ")}); err != nil {
		log.Println(err)
	}
}

// PrintMsg prints message
func (p *Printer) PrintMsg(msg string) {
	type message struct {
		Msg string
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := msgTpl.Execute(p.writer, message{Msg: msg}); err != nil {
		log.Println(err)
	}
}

// PrintMsgErr prints error message
func (p *Printer) PrintMsgErr(msg string) {
	type message struct {
		Msg string
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := msgErrTpl.Execute(p.writer, message{Msg: msg}); err != nil {
		log.Println(err)
	}
}

// PrintRepoErr prints error message
func (p *Printer) PrintRepoErr(msg string, repos []string) {
	type message struct {
		Msg   string
		Repos []string
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := repoErrTpl.Execute(p.writer, message{Msg: msg, Repos: repos}); err != nil {
		log.Println(err)
	}
}

// Print prints result
func (p *Printer) Print(res Result) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := successTpl.Execute(p.writer, res); err != nil {
		log.Println(err)
	}
}

// Error prints error
func (p *Printer) Error(res Result) {
	res.Msg = res.Err.Error()
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := errTpl.Execute(p.errWriter, res); err != nil {
		log.Println(err)
	}
}
