//
// Copyright © 2016 weirdgiraffe <weirdgiraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package gobf

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

type badReader struct {
}

func (r badReader) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("some error")
}

func TestLoadErrorOnReaderError(t *testing.T) {
	r := badReader{}
	p := NewProgram()
	err := p.Load(r)
	if err == nil {
		t.Errorf("No error on Reader Error")
	}
}

func TestResetUsePreviousDataSize(t *testing.T) {
	dataSize := 100
	p := NewProgram()
	p.data = make([]byte, dataSize)
	p.Reset()
	if len(p.data) != dataSize {
		t.Errorf("Previous data size ignored: %v != %v",
			len(p.data), dataSize)
	}
}

type pstate struct {
	cmdInx    int
	cellIndx  int
	cellValue byte
}

func initTestCase(t *testing.T, code string, state pstate) *Program {
	p := NewProgram()
	err := p.Load(strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to load '%v' : %v", code, err)
	}
	p.cmdIndx = state.cmdInx
	p.cellIndx = state.cellIndx
	p.data[p.cellIndx] = state.cellValue
	return p
}

var operationsTestCases = []struct {
	code          string
	initialState  pstate
	expectedState pstate
}{
	{"+", pstate{0, 0, 0}, pstate{1, 0, 1}},
	{"-", pstate{0, 0, 1}, pstate{1, 0, 0}},
	{">", pstate{0, 0, 0}, pstate{1, 1, 0}},
	{"<", pstate{0, 1, 0}, pstate{1, 0, 0}},
	{"[+]", pstate{0, 0, 1}, pstate{1, 0, 1}},
	{"[+]", pstate{0, 0, 0}, pstate{3, 0, 0}},
	{"[+[+]+]", pstate{0, 0, 0}, pstate{7, 0, 0}},
	{"[+]", pstate{2, 0, 0}, pstate{3, 0, 0}},
	{"[+]", pstate{2, 0, 1}, pstate{1, 0, 1}},
	{"[+[+]+]", pstate{6, 0, 1}, pstate{1, 0, 1}},
	{"ignore other symbols", pstate{0, 0, 0}, pstate{1, 0, 0}},
}

func TestOperations(t *testing.T) {
	for _, tt := range operationsTestCases {
		p := initTestCase(t, tt.code, tt.initialState)
		cmdIndx := p.runCmd()

		if cmdIndx != tt.expectedState.cmdInx {
			t.Errorf("Unexpected cmdIndx %v: %v", tt, cmdIndx)
		}
		if p.cellIndx != tt.expectedState.cellIndx {
			t.Errorf("Unexpected cellIndx %v: %v", tt, p.cellIndx)
		}
		cellValue := p.data[p.cellIndx]
		if cellValue != tt.expectedState.cellValue {
			t.Errorf("Unexpected cellValue %v: %v", tt, cellValue)
		}
	}
}

var errorTestCases = []struct {
	code         string
	initialState pstate
}{
	{"<", pstate{0, 0, 0}},
	{"[+", pstate{0, 0, 0}},
	{"+]", pstate{1, 0, 1}},
}

func TestErors(t *testing.T) {
	for _, tt := range errorTestCases {
		p := initTestCase(t, tt.code, tt.initialState)
		err := p.Run()
		if err == nil {
			t.Errorf("No error on %v", tt)
		}
	}
}

func TestPrintCell(t *testing.T) {
	var b bytes.Buffer
	testw := bufio.NewWriter(&b)
	expected := byte('A')
	p := &Program{
		code:     []byte{'.'},
		cmdIndx:  0,
		data:     []byte{expected},
		cellIndx: 0,
		writer:   testw,
	}
	p.runCmd()
	testw.Flush()
	if b.Len() == 0 {
		t.Fatalf("Output buffer is empty")
	}
	if b.Bytes()[0] != expected {
		t.Fatalf("Output mismatch: %v != %v", expected, b.Bytes())
	}
}

func TestScanCell(t *testing.T) {
	expected := []byte("A")
	testr := bytes.NewReader(expected)
	p := &Program{
		code:     []byte{','},
		cmdIndx:  0,
		data:     make([]byte, 1),
		cellIndx: 0,
		reader:   testr,
	}
	p.runCmd()
	if expected[0] != p.currentCell() {
		t.Fatalf("Scan mismatch: %v != %v", expected[0], p.currentCell())
	}
}

func TestSmartDataReallocation(t *testing.T) {
	dataChunkSizeOrig := DataChunkSize
	defer func() { DataChunkSize = dataChunkSizeOrig }()
	DataChunkSize = 2
	p := NewProgram()
	p.Load(strings.NewReader(">>>"))
	p.writer = ioutil.Discard
	err := p.Run()
	if err != nil {
		t.Errorf("Run() return error: %v", err)
	}
	if len(p.data) != 5 {
		t.Errorf("Realloc failed: %v !=5", len(p.data))
	}
}

func TestRunCanReturnError(t *testing.T) {
	p := NewProgram()
	p.Load(strings.NewReader("<"))
	p.writer = ioutil.Discard
	err := p.Run()
	if err == nil {
		t.Error("Run() doenst return errors")
	}
}

func TestRunHelloWorld(t *testing.T) {
	helloWorldText := "++++++++++[>+++++++" +
		">++++++++++>+++>+<<<<-]>++.>+.+++" +
		"++++..+++.>++.<<+++++++++++++++.>" +
		".+++.------.--------.>+.>."
	expected := "Hello World!\n"
	var b bytes.Buffer
	p := NewProgram()
	p.Load(strings.NewReader(helloWorldText))
	bufwriter := bufio.NewWriter(&b)
	p.writer = bufwriter
	err := p.Run()
	if err != nil {
		t.Errorf("Failed to Run 'Hello world' program: %v", err)
	}
	bufwriter.Flush()
	if b.Len() == 0 {
		t.Fatalf("Output buffer is empty")
	}

	if string(b.Bytes()) != expected {
		t.Fatalf("Output mismatch: %v != %v",
			string(b.Bytes()),
			expected)
	}
}

func BenchmarkBeers(b *testing.B) {
	bfBeerText := `99 bottles in 1752 brainfuck instructions
by jim crawford (http://www (dot) goombas (dot) org/)
>++++++++++[<++++++++++>-]<->>>>>+++[>+++>+++<<-]<<<<+<[>[>+
>+<<-]>>[-<<+>>]++++>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<[[-]>>
>>>>[[-]<++++++++++<->>]<-[>+>+<<-]>[<+>-]+>[[-]<->]<<<<<<<<
<->>]<[>+>+<<-]>>[-<<+>>]+>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<
<[>>+>+<<<-]>>>[-<<<+>>>]++>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<
<[>+<[-]]<[>>+<<[-]]>>[<<+>>[-]]<<<[>>+>+<<<-]>>>[-<<<+>>>]+
+++>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<[>+<[-]]<[>>+<<[-]]>>[<
<+>>[-]]<<[[-]>>>++++++++[>>++++++<<-]>[<++++++++[>++++++<-]
>.<++++++++[>------<-]>[<<+>>-]]>.<<++++++++[>>------<<-]<[-
>>+<<]<++++++++[<++++>-]<.>+++++++[>+++++++++<-]>+++.<+++++[
>+++++++++<-]>.+++++..--------.-------.++++++++++++++>>[>>>+
>+<<<<-]>>>>[-<<<<+>>>>]>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<<<
[>>>+>+<<<<-]>>>>[-<<<<+>>>>]+>+<[-<->]<[[-]>>-<<]>>[[-]<<+>
>]<<<[>>+<<[-]]>[>+<[-]]++>>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<
+<[[-]>-<]>[<<<<<<<.>>>>>>>[-]]<<<<<<<<<.>>----.---------.<<
.>>----.+++..+++++++++++++.[-]<<[-]]<[>+>+<<-]>>[-<<+>>]+>+<
[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<<[>>+>+<<<-]>>>[-<<<+>>>]++++
>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<[>+<[-]]<[>>+<<[-]]>>[<<+>
>[-]]<<[[-]>++++++++[<++++>-]<.>++++++++++[>+++++++++++<-]>+
.-.<<.>>++++++.------------.---.<<.>++++++[>+++<-]>.<++++++[
>----<-]>++.+++++++++++..[-]<<[-]++++++++++.[-]]<[>+>+<<-]>>
[-<<+>>]+++>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<[[-]++++++++++.
>+++++++++[>+++++++++<-]>+++.+++++++++++++.++++++++++.------
.<++++++++[>>++++<<-]>>.<++++++++++.-.---------.>.<-.+++++++
++++.++++++++.---------.>.<-------------.+++++++++++++.-----
-----.>.<++++++++++++.---------------.<+++[>++++++<-]>..>.<-
---------.+++++++++++.>.<<+++[>------<-]>-.+++++++++++++++++
.---.++++++.-------.----------.[-]>[-]<<<.[-]]<[>+>+<<-]>>[-
<<+>>]++++>+<[-<->]<[[-]>>-<<]>>[[-]<<+>>]<<[[-]++++++++++.[
-]<[-]>]<+<]`
	p := NewProgram()
	p.Load(strings.NewReader(bfBeerText))
	p.writer = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Run()
		p.Reset()
	}
}
