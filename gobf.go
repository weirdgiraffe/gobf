package gobf

import (
	"fmt"
	"io"
	"io/ioutil"
)

// Program represents brainfuck programm
type Program struct {
	code []byte
	ip   int // next instruction index in code
}

// NewProgram initialize new program
func NewProgram(r io.Reader) *Program {
	code, err := ioutil.ReadAll(r)
	if err != nil {
		panic(fmt.Sprintf("Failed to read program code: %v", err))
	}
	return &Program{
		code: code,
		ip:   -1,
	}
}

// Run runs brainfuck program
func (p *Program) Run() error {
	for p.nextCmd() {
		err := p.runCmd()
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("Not implemented")
}

func (p *Program) nextCmd() bool {
	if p.ip+1 < len(p.code) {
		p.ip++
		return true
	}
	return false
}

func (p *Program) runCmd() error {
	return fmt.Errorf("Bad cmd symbol: '%c' (%v)", p.code[p.ip], p.code[p.ip])
}
