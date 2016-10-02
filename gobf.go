package gobf

import (
	"fmt"
	"io"
	"io/ioutil"
)

// DataChunkSize count of bytes to use when need ot
// increase count of program data cells
var DataChunkSize = 4096

// Program represents brainfuck programm
type Program struct {
	code []byte
	ip   int // current instruction index in code
	data []byte
	dp   int // current data cell index
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
		data: make([]byte, DataChunkSize),
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

func (p *Program) cmd() byte {
	return p.code[p.ip]
}

func (p *Program) cellValue() byte {
	return p.data[p.dp]
}

func (p *Program) runCmd() error {
	switch {
	case p.cmd() == '+':
		return p.incDataCell()
	case p.cmd() == '-':
		return p.decDataCell()
	case p.cmd() == '>':
		return p.incDataPointer()
	case p.cmd() == '<':
		return p.decDataPointer()
	}
	return fmt.Errorf("Bad cmd symbol: '%c' (%v)", p.cmd(), p.cmd())
}

func (p *Program) incDataCell() error {
	if p.cellValue() == 255 {
		return fmt.Errorf("Cell #%d overflow", p.dp)
	}
	p.data[p.dp]++
	return nil
}

func (p *Program) decDataCell() error {
	if p.cellValue() == 0 {
		return fmt.Errorf("Cell #%d underflow", p.dp)
	}
	p.data[p.dp]--
	return nil
}

func (p *Program) incDataPointer() error {
	if p.dp+1 == len(p.data) {
		newData := append(p.data, make([]byte, DataChunkSize)...)
		p.data = newData
	}
	p.dp++
	return nil
}

func (p *Program) decDataPointer() error {
	if p.dp == 0 {
		return fmt.Errorf("Data pointer underfow")
	}
	p.dp--
	return nil
}
