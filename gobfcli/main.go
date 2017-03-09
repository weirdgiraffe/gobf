//
// Copyright © 2016 weirdgiraffe <weirdgiraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"fmt"
	"os"

	"github.com/weirdgiraffe/gobf"
)

func main() {
	p := gobf.NewProgram()
	err := p.Load(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
