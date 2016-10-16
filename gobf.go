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
func (p *Program) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	for p.cmdIndx < len(p.code) {
		p.runCmd()
	}
	return nil
}

func (p *Program) runCmd() {
	switch p.code[p.cmdIndx] {
	case '+':
		p.data[p.cellIndx]++
	case '-':
		p.data[p.cellIndx]--
	case '>':
		if p.cellIndx+1 > len(p.data) {
			p.data = append(p.data, make([]uint8, DataChunkSize)...)
		}
		p.cellIndx++
	case '<':
		p.cellIndx--
	case '[':
		if p.data[p.cellIndx] == 0 {
			for depth := 1; depth > 0; {
				p.cmdIndx++
				switch p.code[p.cmdIndx] {
				case '[':
					depth++
				case ']':
					depth--
				}
			}
		}
	case ']':
		if p.data[p.cellIndx] != 0 {
			for depth := 1; depth > 0; {
				p.cmdIndx--
				switch p.code[p.cmdIndx] {
				case ']':
					depth++
				case '[':
					depth--
				}
			}
		}
	case '.':
		_, err := p.writer.Write([]byte{p.data[p.cellIndx]})
		if err != nil {
			panic(err)
		}
	case ',':
		var b = make([]byte, 1)
		_, err := p.reader.Read(b)
		if err != nil {
			panic(err)
		}
		p.data[p.cellIndx] = b[0]
	}
	p.cmdIndx++
}
