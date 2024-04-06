# mooltah - Go Template Rendition Cli

mooltah is a pure Go based CLI which renders the [go-templates](https://pkg.go.dev/text/template) by hierarchical configurations.
This cli is based on python Jinja2 template renderer program [yasha](https://github.com/kblomqvist/yasha).
I dedicate this cli to my GrandFather Late Shri Moolchand ji Jat.

## Installation

* The code can be compiled using Golang 1.21+ 

```bash
make build
```

This will create executable file `main`, which can be copied to any location in your `$PATH`

```bash
make install
```

## Usage

```bash
mooltah --help
this program renders input template to output with the configurations given by --variable files
mooltah 1.0.0
Usage: mooltah [--variable VARIABLE] [--output OUTPUT] [INPUTTEMPLATEFILE]

Positional arguments:
  INPUTTEMPLATEFILE      Template File which will be rendered to OUTPUT

Options:
  --variable VARIABLE, -v VARIABLE
                         Read variables from YAML, JSON, TOML, and/or Key=Value Files
  --output OUTPUT, -o OUTPUT
                         Output file which will have rendition of input template file
  --help, -h             display this help and exit
  --version              display version and exit

```

## Sample Run

* Sample run while having hierarchical configurations read from

```bash
mooltah -v demo/file.yaml -v demo/file.json -v demo/file.kv -o demo/out.yaml demo/template.go ; cat demo/out.yaml
```

```
2024/04/06 18:39:51 INFO demo/template.go
map[a:1.2 b:map[c:3 d:[1 4 9]] e:7 h:123 z:f=m]
a: 1.2
h: 123
b:

  3

  [1 4 9]

```

## Acknowledgement

I would like to acknowledge to the creator(s) and maintainers of the wonderful tool [yasha](https://github.com/kblomqvist/yasha)

