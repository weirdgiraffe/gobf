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

// DataChunkSize count of program cells
var DataChunkSize = 30000

const (
	MODIFY = iota
	SHIFT
	PRINT
	SCAN
	BEGINLOOP
	ENDLOOP
	CLEAR
	ADD
	SUB
)

type instruction struct {
	itype int
	iarg  int
}

func compileFlat(text []byte) (ret []instruction, last int) {
	for i, c := range text {
		switch c {
		case '+':
			ret = append(ret, instruction{MODIFY, 1})
		case '-':
			ret = append(ret, instruction{MODIFY, -1})
		case '>':
			ret = append(ret, instruction{SHIFT, 1})
		case '<':
			ret = append(ret, instruction{SHIFT, -1})
		case '.':
			ret = append(ret, instruction{PRINT, 1})
		case ',':
			ret = append(ret, instruction{SCAN, 1})
		case '[', ']':
			return ret, i
		}
	}
	return ret, len(text)
}

func _optimizeFlat(flat []instruction) (ret []instruction) {
	ci := 0
	value := flat[ci].iarg
	for i := 1; i < len(flat); i++ {
		if flat[ci].itype == flat[i].itype {
			value += flat[i].iarg
		} else {
			if value != 0 {
				flat[ci].iarg = value
				ret = append(ret, flat[ci])
			}
			ci = i
			value = flat[ci].iarg
		}
	}
	if value != 0 {
		flat[ci].iarg = value
		ret = append(ret, flat[ci])
	}
	return ret
}

func optimizeFlat(flat []instruction) []instruction {
	if len(flat) > 0 {
		return _optimizeFlat(flat)
	}
	return flat
}

func isLoopFlat(loop []instruction) bool {
	loopBody := loop[1 : len(loop)-1]
	for _, i := range loopBody {
		switch i.itype {
		case BEGINLOOP, ENDLOOP:
			return false
		}
	}
	return true
}

func optimizeFlatLoop(loop []instruction) (ret []instruction) {
	loopBody := loop[1 : len(loop)-1]
	ret = append(ret, loop[0])
	ret = append(ret, optimizeFlat(loopBody)...)
	ret = append(ret, loop[len(loop)-1])
	return ret
}

func isEmptyLoop(loop []instruction) bool {
	if len(loop) == 2 {
		return true
	}
	return false
}

func optimizeClearLoop(loop []instruction) ([]instruction, bool) {
	loopBody := loop[1 : len(loop)-1]
	if len(loop) == 2 {
		return loop, false
	}
	onlyModifications := true
	for _, i := range loopBody {
		switch i.itype {
		default:
			onlyModifications = false
		case MODIFY:
		}
	}
	if onlyModifications {
		return []instruction{instruction{CLEAR, 1}}, true
	}
	return loop, false
}

func optimizeAddLoop(loop []instruction) ([]instruction, bool) {
	loopBody := loop[1 : len(loop)-1]
	if len(loopBody) == 4 {
		// SHIT MODIFY SHIFT MODIFY
		if loopBody[0].itype == SHIFT &&
			loopBody[1].itype == MODIFY &&
			loopBody[2].itype == SHIFT &&
			loopBody[3].itype == MODIFY &&
			loopBody[0].iarg == -loopBody[2].iarg &&
			loopBody[3].iarg == -1 {
			if loopBody[1].iarg == 1 {
				return []instruction{{ADD, loopBody[0].iarg}}, true
			} else if loopBody[1].iarg == -1 {
				return []instruction{{SUB, loopBody[0].iarg}}, true
			}
		}
		// OR MODIFY SHIFT MODIFY SHIFT
		if loopBody[0].itype == MODIFY &&
			loopBody[1].itype == SHIFT &&
			loopBody[2].itype == MODIFY &&
			loopBody[3].itype == SHIFT &&
			loopBody[1].iarg == -loopBody[3].iarg &&
			loopBody[0].iarg == -1 {
			if loopBody[2].iarg == 1 {
				return []instruction{{ADD, loopBody[1].iarg}}, true
			} else if loopBody[2].iarg == -1 {
				return []instruction{{SUB, loopBody[1].iarg}}, true
			}
		}
	}
	return loop, false
}

func optimizeLoop(loop []instruction) []instruction {
	if isLoopFlat(loop) {
		loop = optimizeFlatLoop(loop)
		if ret, ok := optimizeClearLoop(loop); ok {
			return ret
		}
		if ret, ok := optimizeAddLoop(loop); ok {
			return ret
		}
	}
	return loop
}

func compile(text []byte) (ret []instruction, last int) {
	for i := 0; i < len(text); i++ {
		flat, idiff := compileFlat(text[i:])
		flat = optimizeFlat(flat)
		i += idiff
		if len(flat) > 0 {
			ret = append(ret, flat...)
		}
		if i < len(text) {
			switch text[i] {
			case '[':
				loop, idiff := compile(text[i+1:])
				ll := len(loop) + 1
				loop = append([]instruction{{BEGINLOOP, ll}}, loop...)
				loop = append(loop, instruction{ENDLOOP, ll})
				loop = optimizeLoop(loop)
				if isEmptyLoop(loop) {
					panic(fmt.Errorf("empty loop found"))
				}
				ret = append(ret, loop...)
				i += idiff
			case ']':
				return ret, i + 1
			}
		}
	}
	return ret, len(text)
}

// Program represents brainfuck programm
type Program struct {
	code   []instruction
	reader io.Reader
	writer io.Writer
}

// NewProgram initialize empty program
func NewProgram() *Program {
	return &Program{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

// Load load program code
func (p *Program) Load(r io.Reader) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	text, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	p.code, _ = compile(text)
	return err
}

// Run runs brainfuck program
func (p *Program) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	iobyte := make([]byte, 1)

	cells := make([]byte, DataChunkSize)
	cellIndx := 0
	for i := 0; i < len(p.code); i++ {
		cell := &cells[cellIndx]
		switch p.code[i].itype {
		case CLEAR:
			*cell = 0
		case ADD:
			cells[cellIndx+p.code[i].iarg] += *cell
			*cell = 0
		case SUB:
			cells[cellIndx+p.code[i].iarg] -= *cell
			*cell = 0
		case MODIFY:
			*cell += byte(p.code[i].iarg)
		case SHIFT:
			cellIndx += p.code[i].iarg
		case BEGINLOOP:
			if *cell == 0 {
				i += p.code[i].iarg
			}
		case ENDLOOP:
			if *cell != 0 {
				i -= p.code[i].iarg
			}
		case PRINT:
			iobyte[0] = *cell
			for repeat := 0; repeat < p.code[i].iarg; repeat++ {
				_, err := p.writer.Write(iobyte)
				if err != nil {
					panic(err)
				}
			}
		case SCAN:
			_, err := p.reader.Read(cells[cellIndx : cellIndx+1])
			if err != nil {
				panic(err)
			}
		}
	}
	return err
}
