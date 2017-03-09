# Brainfuck interpreter in Go [![GoDoc](https://godoc.org/github.com/weirdgiraffe/gobf?status.svg)](https://godoc.org/github.com/weirdgiraffe/gobf) [![CircleCI](https://circleci.com/gh/weirdgiraffe/gobf.svg?style=shield)](https://circleci.com/gh/weirdgiraffe/gobf) [![Coverage Status](https://coveralls.io/repos/github/weirdgiraffe/gobf/badge.svg?branch=master)](https://coveralls.io/github/weirdgiraffe/gobf?branch=master)

## And there is some comparison of brainfuck clients [here](https://github.com/weirdgiraffe/bfbench)

## Interpreter has a commandline client: **gobfcli**


`gobfcli` simply read `stdin`, interpret Brainfuck and do output to `stdout`

Install:

    go get github.com/weirdgiraffe/gobf/gobfcli

Usage:

    gobfcli < someprogram.bf

or

    curl https://copy.sh/brainfuck/prog/mandelbrot.b | gobfcli

Worth to mention:

- [Beautiful brainfuck visualisation](https://fatiherikli.github.io/brainfuck-visualizer/)
- [Relative big collection of brainfuck programs + interpreter in JS](https://copy.sh/brainfuck/)

