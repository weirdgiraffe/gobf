//
// Copyright © 2016 weirdgiraffe <weirdgiraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package gobf

import (
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
	prevbo   int
	prevbc   int
	reader   io.Reader
	writer   io.Writer
}

// NewProgram initialize empty program
func NewProgram() *Program {
	return &Program{
		reader: os.Stdin,
		writer: os.Stdout,
		prevbc: -1,
		prevbo: -1,
	}
}

// Load load program code
func (p *Program) Load(r io.Reader) (err error) {
	code, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	p.code = code
	p.data = make([]byte, DataChunkSize)
	p.Reset()
	return err
}

// Reset resets program. Run() will run program again
func (p *Program) Reset() {
	for i := range p.data {
		p.data[i] = 0
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
	return err
}

func (p *Program) runCmd() {
	switch p.code[p.cmdIndx] {
	case '+':
		p.data[p.cellIndx]++
	case '-':
		p.data[p.cellIndx]--
	case '>':
		if p.cellIndx+1 > len(p.data) {
			p.data = append(p.data, make([]byte, DataChunkSize)...)
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
			bc := p.prevbc
			p.prevbc = p.cmdIndx
			if bc == p.cmdIndx && p.prevbo != -1 {
				p.cmdIndx = p.prevbo + 1
				return
			}
			for depth := 1; depth > 0; {
				p.cmdIndx--
				switch p.code[p.cmdIndx] {
				case ']':
					depth++
				case '[':
					depth--
				}
			}
			p.prevbo = p.cmdIndx
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
