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
