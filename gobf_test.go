package gobf

import (
	"strings"
	"testing"
)

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
	cmd           byte
	initialValue  byte
	expectedValue byte
	expectedError bool
}{
	{'+', 0, 1, false},
	{'+', 255, 255, true},
	{'-', 1, 0, false},
	{'-', 0, 0, true},
}

func TestCellValueOperations(t *testing.T) {
	for _, tt := range cellValueOperationsTests {
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

func TestRunHelloWorld(t *testing.T) {
	helloWorldText := "++++++++++[>+++++++" +
		">++++++++++>+++>+<<<<-]>++.>+.+++" +
		"++++..+++.>++.<<+++++++++++++++.>" +
		".+++.------.--------.>+.>."
	p := NewProgram(strings.NewReader(helloWorldText))
	err := p.Run()
	if err != nil {
		t.Errorf("Failed to Run 'Hello world' program: %v", err)
	}
}
