package gobf

import (
	"fmt"
	"io"
	"io/ioutil"
)

// Program represents brainfuck programm
type Program struct {
	code []byte
}

// NewProgram initialize new program
func NewProgram(r io.Reader) *Program {
	code, err := ioutil.ReadAll(r)
	if err != nil {
		panic(fmt.Sprintf("Failed to read program code: %v", err))
	}
	return &Program{
		code: code,
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
	return false
}

func (p *Program) runCmd() error {
	return fmt.Errorf("Bad cmd symbol: %v", p.cmd)
}
