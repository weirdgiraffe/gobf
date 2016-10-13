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

func TestNewProgramPanicOnReaderError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("No panic on Reader Error")
		}
	}()
	r := badReader{}
	_ = NewProgram(r)
}

var nextCmdTests = []struct {
	code     []byte
	ip       int
	expected bool
}{
	{[]byte{'+'}, -1, true},
	{[]byte{'+'}, 0, false},
	{[]byte{}, -1, false},
	{[]byte{}, 0, false},
}

func TestNextCmd(t *testing.T) {
	for _, tt := range nextCmdTests {
		p := &Program{code: tt.code, ip: tt.ip}
		if ok := p.nextCmd(); ok != tt.expected {
			t.Errorf("code:'%v' ip:%d nextCmd(): %v != %v",
				tt.code, tt.ip, tt.expected, ok)
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
		p := &Program{
			code: []byte{tt.cmd},
			data: make([]byte, 1),
		}
		p.data[0] = tt.initialValue
		err := p.runCmd()
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%v error: %v", tt, err)
			}
		}
		if tt.expectedValue != p.cellValue() {
			t.Errorf("%c %v value mismatch: %v != %v",
				tt.cmd, tt, p.cellValue(), tt.expectedValue)
		}
	}
}

var dataPointerOperationsTests = []struct {
	cmd           byte
	initialDp     int
	expectedDp    int
	expectedError bool
}{
	{'>', 0, 1, false},
	{'>', DataChunkSize, DataChunkSize + 1, false},
	{'<', 1, 0, false},
	{'<', 0, 0, true},
}

func TestDataPointerOperations(t *testing.T) {
	for _, tt := range dataPointerOperationsTests {
		p := &Program{
			code: []byte{tt.cmd},
			data: make([]byte, DataChunkSize),
			dp:   tt.initialDp,
		}
		err := p.runCmd()
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%c %v error: %v", tt.cmd, tt, err)
			}
		}
		if tt.expectedDp != p.dp {
			t.Errorf("%c %v value mismatch: %v != %v",
				tt.cmd, tt, p.dp, tt.expectedDp)
		}
	}
}

var moveCpOperationsTests = []struct {
	code             string
	initialIp        int
	initialCellValue byte
	expectedIp       int
	expectedError    bool
}{
	{"[+]", 0, 0, 2, false},
	{"[+]", 0, 1, 0, false},
	{"[++", 0, 0, 2, true},
	{"[+[++]+]", 0, 0, 7, false},
	{"[+]", 2, 1, 0, false},
	{"[+]", 2, 0, 2, false},
	{"++]", 2, 1, 0, true},
	{"[+[++]+]", 7, 1, 0, false},
}

func TestMoveCpOperations(t *testing.T) {
	for _, tt := range moveCpOperationsTests {
		p := &Program{
			code: []byte(tt.code),
			ip:   tt.initialIp,
			data: make([]byte, 10),
		}
		p.data[p.dp] = tt.initialCellValue
		err := p.runCmd()
		if err != nil {
			if tt.expectedError == false {
				t.Fatalf("%v error: %v", tt, err)
			}
		}
		if tt.expectedIp != p.ip {
			t.Errorf("%v value mismatch: %v != %v",
				tt, p.ip, tt.expectedIp)
		}
	}
}

func TestPrintCell(t *testing.T) {
	var b bytes.Buffer
	testw := bufio.NewWriter(&b)
	expected := byte('A')
	p := &Program{
		code:   []byte("."),
		data:   []byte{expected},
		writer: testw,
	}
	err := p.runCmd()
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
		code:   []byte(","),
		data:   make([]byte, 1),
		reader: testr,
	}
	err := p.runCmd()
	if err != nil {
		t.Fatalf("Failed to scan cell value")
	}
	if expected[0] != p.cellValue() {
		t.Fatalf("Scan mismatch: %v != %v", expected[0], p.cellValue())
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
		p := NewProgram(strings.NewReader(tt.code))
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
	p := NewProgram(strings.NewReader(helloWorldText))
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
	p := NewProgram(strings.NewReader(bfBeerText))
	p.writer = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Run()
		p.Reset()
	}
}
