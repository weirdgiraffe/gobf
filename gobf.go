package gobf

import (
	"fmt"
	"io"
)

type Program struct {
}

func NewProgram(r io.Reader) *Program {
	return &Program{}
}

func (p *Program) Run() error {
	return fmt.Errorf("Not implemented")
}
