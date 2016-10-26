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

func TestLoadErrorOnEmptyLoop(t *testing.T) {
	p := NewProgram()
	err := p.Load(strings.NewReader("[+-]"))
	if err == nil {
		t.Errorf("No error on empty loop")
	}
}

func iEqual(a, b []instruction) bool {
	if a == nil && b == nil {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

var compileCases = []struct {
	text     string
	expected []instruction
}{
	{"+", []instruction{{MODIFY, 1}}},
	{"++", []instruction{{MODIFY, 2}}},
	{"-", []instruction{{MODIFY, -1}}},
	{"--", []instruction{{MODIFY, -2}}},
	{"+-", []instruction{}},
	{">", []instruction{{SHIFT, 1}}},
	{"<", []instruction{{SHIFT, -1}}},
	{".", []instruction{{PRINT, 1}}},
	{",", []instruction{{SCAN, 1}}},
	{
		"[>++<-]",
		[]instruction{
			{BEGINLOOP, 5},
			{SHIFT, 1},
			{MODIFY, 2},
			{SHIFT, -1},
			{MODIFY, -1},
			{ENDLOOP, 5},
		},
	},
	{
		"[>[+]<-]",
		[]instruction{
			{BEGINLOOP, 5},
			{SHIFT, 1},
			{CLEAR, 1},
			{SHIFT, -1},
			{MODIFY, -1},
			{ENDLOOP, 5},
		},
	},
	{
		"[>[++-]<-]",
		[]instruction{
			{BEGINLOOP, 5},
			{SHIFT, 1},
			{CLEAR, 1},
			{SHIFT, -1},
			{MODIFY, -1},
			{ENDLOOP, 5},
		},
	},
	{
		"[+[->++<]<-]",
		[]instruction{
			{BEGINLOOP, 10},
			{MODIFY, 1},
			{BEGINLOOP, 5},
			{MODIFY, -1},
			{SHIFT, 1},
			{MODIFY, 2},
			{SHIFT, -1},
			{ENDLOOP, 5},
			{SHIFT, -1},
			{MODIFY, -1},
			{ENDLOOP, 10},
		},
	},
	{
		"[->>>+<<<]",
		[]instruction{
			{ADD, 3},
		},
	},
	{
		"[>>+<<-]",
		[]instruction{
			{ADD, 2},
		},
	},
}

func TestCompilation(t *testing.T) {
	for _, tt := range compileCases {
		i, _ := compile([]byte(tt.text))
		if iEqual(i, tt.expected) == false {
			t.Errorf("mismatch '%v':\n%v\n!=\n%v",
				tt.text, i, tt.expected)
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

var bfBeerText = `99 bottles in 1752 brainfuck instructions
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

func BenchmarkBeerCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = compile([]byte(bfBeerText))
	}
}

func TestRunBeer(t *testing.T) {
	var expected string
	for i := 99; i > 2; i-- {
		expected += fmt.Sprintf("%d Bottles of beer on the wall\n", i)
		expected += fmt.Sprintf("%d Bottles of beer\n", i)
		expected += "Take one down and pass it around\n"
		expected += fmt.Sprintf("%d Bottles of beer on the wall\n\n", i-1)
	}
	expected += "2 Bottles of beer on the wall\n"
	expected += "2 Bottles of beer\n"
	expected += "Take one down and pass it around\n"
	expected += "1 Bottle of beer on the wall\n\n"
	expected += "1 Bottle of beer on the wall\n"
	expected += "1 Bottle of beer\n"
	expected += "Take one down and pass it around\n"
	expected += "0 Bottles of beer on the wall\n\n"

	var b bytes.Buffer
	p := NewProgram()
	p.Load(strings.NewReader(bfBeerText))
	bufwriter := bufio.NewWriter(&b)
	p.writer = bufwriter
	err := p.Run()
	if err != nil {
		t.Errorf("Failed to Run 'Beer' program: %v", err)
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

func BenchmarkBeerRun(b *testing.B) {
	p := NewProgram()
	p.Load(strings.NewReader(bfBeerText))
	//fmt.Println(p.code)
	p.writer = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := p.Run()
		if err != nil {
			b.Error(err)
		}
	}
}
