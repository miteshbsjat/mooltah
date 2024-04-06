# mooltah - Go Template Rendition Cli

mooltah is a pure Go based CLI which renders the [go-templates](https://pkg.go.dev/text/template) by hierarchical configurations.
This cli is based on python Jinja2 template renderer program [yasha](https://github.com/kblomqvist/yasha).
I dedicate this cli to my GrandFather Late Shri Moolchand ji Jat.

## Installation

* The code can be compiled using Golang 1.21+ 

```bash
go build main.go
```

This will create executable file `main`, which can be copied to any location in your `$PATH`

```bash
sudo cp main /usr/local/bin/mooltah
```

## Usage

```bash
 command-line tool for processing YAML, JSON, and TOML files

Usage:
  mooltah [flags] <input_template_file>

Flags:
  -v, --files strings   List of files to process
  -h, --help            help for mooltah
  -o, --output string   Output file
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

