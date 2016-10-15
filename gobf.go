//
// Copyright © 2016 weirdgiraffe <weirdgiraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package gobf

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// DataChunkSize count of bytes to use when need ot
// increase count of program data cells
var DataChunkSize = 30000

// AllowOverflows do allow overflow (255+1) and underflow (0-1) of a cell value
var AllowOverflows = true

// Program represents brainfuck programm
type Program struct {
	code     []byte
	cmdIndx  int
	data     []byte
	cellIndx int
	reader   io.Reader
	writer   io.Writer
}

// NewProgram initialize empty program
func NewProgram() *Program {
	return &Program{
		code:    []byte{},
		cmdIndx: -1,
		reader:  os.Stdin,
		writer:  os.Stdout,
	}
}

// Load load program code
func (p *Program) Load(r io.Reader) error {
	code, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("Failed to read program code: %v", err)
	}
	p.code = code
	p.Reset()
	return nil
}

// Reset resets program. RUn() will run program again
func (p *Program) Reset() {
	p.data = make([]byte, DataChunkSize)
	p.cellIndx = 0
	p.cmdIndx = -1
}

// Run runs brainfuck program
func (p *Program) Run() error {
	for p.nextCmd() {
		err := p.runCmd()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Program) nextCmd() bool {
	if p.cmdIndx+1 < len(p.code) {
		p.cmdIndx++
		return true
	}
	return false
}

func (p *Program) cmd(indx int) byte {
	return p.code[indx]
}

func (p *Program) currentCell() byte {
	return p.data[p.cellIndx]
}

func (p *Program) runCmd() error {
	return p._runCmd(p.code[p.cmdIndx])
}

func (p *Program) _runCmd(cmd byte) error {
	switch {
	case cmd == '+':
		return p.cmdIncCell()
	case cmd == '-':
		return p.cmdDecCell()
	case cmd == '>':
		return p.cmdNextCell()
	case cmd == '<':
		return p.cmdPrevCell()
	case cmd == '[':
		return p.cmdForward()
	case cmd == ']':
		return p.cmdBackward()
	case cmd == '.':
		return p.cmdPrintCell()
	case cmd == ',':
		return p.cmdScanCell()
	}
	return nil // simply ignore all unused symbols
}

func (p *Program) cmdIncCell() error {
	if p.currentCell() == 255 && AllowOverflows == false {
		return fmt.Errorf(
			"Cell #%d overflow (offset: %d)",
			p.cellIndx, p.cmdIndx)
	}
	p.data[p.cellIndx]++
	return nil
}

func (p *Program) cmdDecCell() error {
	if p.currentCell() == 0 && AllowOverflows == false {
		return fmt.Errorf(
			"Cell #%d underflow (offset: %d)",
			p.cellIndx, p.cmdIndx)
	}
	p.data[p.cellIndx]--
	return nil
}

func (p *Program) cmdNextCell() error {
	if p.cellIndx+1 >= len(p.data) {
		newData := append(p.data, make([]byte, DataChunkSize)...)
		p.data = newData
	}
	p.cellIndx++
	return nil
}

func (p *Program) cmdPrevCell() error {
	if p.cellIndx == 0 {
		return fmt.Errorf("Data pointer underfow")
	}
	p.cellIndx--
	return nil
}

func (p *Program) _cmdForward() error {
	for seen, i := 0, p.cmdIndx+1; i < len(p.code); i++ {
		switch {
		case p.cmd(i) == '[':
			seen++
		case p.cmd(i) == ']':
			if seen == 0 {
				p.cmdIndx = i
				return nil
			}
			seen--
		}
	}
	return fmt.Errorf("No closing ']' found")
}

func (p *Program) cmdForward() error {
	// if current cell value is 0,
	// increase cmdIndx until matching bracket
	if p.currentCell() != 0 {
		return nil
	}
	return p._cmdForward()
}

func (p *Program) _cmdBackward() error {
	for seen, i := 0, p.cmdIndx-1; i >= 0; i-- {
		switch {
		case p.cmd(i) == ']':
			seen++
		case p.cmd(i) == '[':
			if seen == 0 {
				p.cmdIndx = i
				return nil
			}
			seen--
		}
	}
	return fmt.Errorf("No closing '[' found")
}

func (p *Program) cmdBackward() error {
	// if current cell value is not 0,
	// decrease cmdIndx until matching bracket
	if p.currentCell() == 0 {
		return nil
	}
	return p._cmdBackward()
}

func (p *Program) cmdPrintCell() error {
	_, err := p.writer.Write([]byte{p.currentCell()})
	return err
}

func (p *Program) cmdScanCell() error {
	var b = make([]byte, 1)
	_, err := p.reader.Read(b)
	if err == nil {
		p.data[p.cellIndx] = b[0]
	}
	return err
}
