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
		cmdIndx: 0,
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

// Reset resets program. Run() will run program again
func (p *Program) Reset() {
	if len(p.data) > 0 {
		p.data = make([]byte, len(p.data))
	} else {
		p.data = make([]byte, DataChunkSize)
	}

	p.cellIndx = 0
	p.cmdIndx = 0
}

// Run runs brainfuck program
func (p *Program) Run() error {
	var err error
	for p.cmdIndx < len(p.code) {
		p.cmdIndx, err = p.runCmd()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Program) runCmd() (int, error) {
	switch p.code[p.cmdIndx] {
	default:
		return p.cmdIndx + 1, nil
	case '+':
		return p.cmdIncCellValue()
	case '-':
		return p.cmdDecCellValue()
	case '>':
		return p.cmdNextCell()
	case '<':
		return p.cmdPrevCell()
	case '[':
		return p.cmdForward()
	case ']':
		return p.cmdBackward()
	case '.':
		return p.cmdPrintCell()
	case ',':
		return p.cmdScanCell()
	}
}

func (p *Program) cmd(indx int) byte {
	return p.code[indx]
}

func (p *Program) currentCell() byte {
	return p.data[p.cellIndx]
}

func (p *Program) opcount(op byte) int {
	for i, c := range p.code[p.cmdIndx:] {
		if c != op {
			return i
		}
	}
	return len(p.code)
}

func (p *Program) cmdIncCellValue() (int, error) {
	count := p.opcount('+')
	p.data[p.cellIndx] += byte(count)
	return p.cmdIndx + count, nil
}

func (p *Program) cmdDecCellValue() (int, error) {
	count := p.opcount('-')
	p.data[p.cellIndx] -= byte(count)
	return p.cmdIndx + count, nil
}

func (p *Program) cmdNextCell() (int, error) {
	count := p.opcount('>')
	if p.cellIndx+count >= len(p.data) {
		incSize := (count / DataChunkSize) + DataChunkSize
		p.data = append(p.data, make([]byte, incSize)...)
	}
	p.cellIndx += count
	return p.cmdIndx + count, nil
}

func (p *Program) cmdPrevCell() (int, error) {
	count := p.opcount('<')
	if p.cellIndx-count < 0 {
		return 0, fmt.Errorf("Data pointer underfow")
	}
	p.cellIndx -= count
	return p.cmdIndx + count, nil
}

func (p *Program) _cmdForward() (int, error) {
	for seen, i := 0, p.cmdIndx+1; i < len(p.code); i++ {
		switch p.cmd(i) {
		case '[':
			seen++
		case ']':
			if seen == 0 {
				return i + 1, nil
			}
			seen--
		}
	}
	return len(p.code), fmt.Errorf("No closing ']' found")
}

func (p *Program) cmdForward() (int, error) {
	// if current cell value is 0,
	// increase cmdIndx until matching bracket
	if p.currentCell() != 0 {
		return p.cmdIndx + 1, nil
	}
	return p._cmdForward()
}

func (p *Program) _cmdBackward() (int, error) {
	for seen, i := 0, p.cmdIndx-1; i >= 0; i-- {
		switch p.cmd(i) {
		case ']':
			seen++
		case '[':
			if seen == 0 {
				return i + 1, nil
			}
			seen--
		}
	}
	return len(p.code), fmt.Errorf("No closing '[' found")
}

func (p *Program) cmdBackward() (int, error) {
	// if current cell value is not 0,
	// decrease cmdIndx until matching bracket
	if p.currentCell() == 0 {
		return p.cmdIndx + 1, nil
	}
	return p._cmdBackward()
}

func (p *Program) cmdPrintCell() (int, error) {
	_, err := p.writer.Write([]byte{p.currentCell()})
	return p.cmdIndx + 1, err
}

func (p *Program) cmdScanCell() (int, error) {
	var b = make([]byte, 1)
	_, err := p.reader.Read(b)
	if err == nil {
		p.data[p.cellIndx] = b[0]
	}
	return p.cmdIndx + 1, err
}
