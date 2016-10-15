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

var nextCmdTests = []struct {
	code     []byte
	cmdIndx  int
	expected bool
}{
	{[]byte{'+'}, -1, true},
	{[]byte{'+'}, 0, false},
	{[]byte{}, -1, false},
	{[]byte{}, 0, false},
}

func TestNextCmd(t *testing.T) {
	for _, tt := range nextCmdTests {
		p := &Program{code: tt.code, cmdIndx: tt.cmdIndx}
		if ok := p.nextCmd(); ok != tt.expected {
			t.Errorf("code:'%v' ip:%d nextCmd(): %v != %v",
				tt.code, tt.cmdIndx, tt.expected, ok)
		}
	}
}

var cellValueOperationsTests = []struct {
	cmd            byte
	initialValue   byte
	allowOverflows bool
	expectedValue  byte
	expectedError  bool
}{
	{'+', 0, false, 1, false},
	{'+', 255, true, 0, false},
	{'+', 255, false, 255, true},
	{'-', 1, false, 0, false},
	{'-', 0, false, 0, true},
	{'-', 0, true, 255, false},
}

func TestCellValueOperations(t *testing.T) {
	defer func() { AllowOverflows = true }()
	for _, tt := range cellValueOperationsTests {
		AllowOverflows = tt.allowOverflows
		p := &Program{data: make([]byte, 1)}
		p.data[0] = tt.initialValue
		err := p._runCmd(tt.cmd)
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%v error: %v", tt, err)
			}
		}
		if tt.expectedValue != p.currentCell() {
			t.Errorf("%c %v value mismatch: %v != %v",
				tt.cmd, tt, p.currentCell(), tt.expectedValue)
		}
	}
}

var dataPointerOperationsTests = []struct {
	cmd              byte
	initialCellIndx  int
	expectedCellIndx int
	expectedError    bool
}{
	{'>', 0, 1, false},
	{'>', DataChunkSize, DataChunkSize + 1, false},
	{'<', 1, 0, false},
	{'<', 0, 0, true},
}

func TestDataPointerOperations(t *testing.T) {
	for _, tt := range dataPointerOperationsTests {
		p := &Program{
			data:     make([]byte, DataChunkSize),
			cellIndx: tt.initialCellIndx,
		}
		err := p._runCmd(tt.cmd)
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%c %v error: %v", tt.cmd, tt, err)
			}
		}
		if tt.expectedCellIndx != p.cellIndx {
			t.Errorf("%c %v value mismatch: %v != %v",
				tt.cmd, tt, p.cellIndx, tt.expectedCellIndx)
		}
	}
}

var moveCpOperationsTests = []struct {
	code             string
	initialCmdIndx   int
	initialCellValue byte
	expectedCmdIndx  int
	expectedError    bool
}{
	{"[+]", 0, 0, 2, false},
	{"[+]", 0, 1, 0, false},
	{"[++", 0, 0, 0, true},
	{"[+[++]+]", 0, 0, 7, false},
	{"[+]", 2, 1, 0, false},
	{"[+]", 2, 0, 2, false},
	{"++]", 2, 1, 2, true},
	{"[+[++]+]", 7, 1, 0, false},
}

func TestMoveCpOperations(t *testing.T) {
	for _, tt := range moveCpOperationsTests {
		p := NewProgram()
		p.Reset()
		p.code = []byte(tt.code)
		p.cmdIndx = tt.initialCmdIndx
		p.data[0] = tt.initialCellValue
		err := p.runCmd()
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%v error: %v", tt, err)
			}
		}
		if tt.expectedCmdIndx != p.cmdIndx {
			t.Errorf("%v value mismatch: %v != %v",
				tt, p.cmdIndx, tt.expectedCmdIndx)
		}
	}
}

func TestPrintCell(t *testing.T) {
	var b bytes.Buffer
	testw := bufio.NewWriter(&b)
	expected := byte('A')
	p := &Program{
		data:   []byte{expected},
		writer: testw,
	}
	err := p._runCmd('.')
	if err != nil {
		t.Fatalf("Failed to print cell value")
	}
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
		data:   make([]byte, 1),
		reader: testr,
	}
	err := p._runCmd(',')
	if err != nil {
		t.Fatalf("Failed to scan cell value")
	}
	if expected[0] != p.currentCell() {
		t.Fatalf("Scan mismatch: %v != %v", expected[0], p.currentCell())
	}
}

var errorHandlingTests = []struct {
	code          string
	expectedError bool
}{
	{".++>.++HELLO WORLD++.<++.", false},
	{"><<", true},
}

func TestRunErrorHadling(t *testing.T) {
	for _, tt := range errorHandlingTests {
		p := NewProgram()
		p.Load(strings.NewReader(tt.code))
		err := p.Run()
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%v error: %v", tt, err)
			}
		}
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
