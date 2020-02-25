package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var cfg struct {
	Output flags.Filename `short:"o" long:"out" description:"output file path" default:"-"`
	Append bool           `short:"a" long:"append" description:"append to output file"`

	Dirs []string `short:"d" long:"dir" description:"directories to parse"`

	Package string   `short:"p" long:"pkg" description:"package that the files belong to"`
	Args struct {
		Files   []string `positional-arg-name:"go-source-file"`
	} `positional-args:"yes"`
}

func main() {
	var parser = flags.NewParser(&cfg, flags.Default)
	parser.ShortDescription = "Golang constants parser"
	parser.LongDescription = "Parse go source files and export constants definitions"
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Println(parser.Usage)
			os.Exit(0)
		} else {
			fmt.Println("Failed to parse arguments:", err)
			os.Exit(1)
		}
	}

	file := os.Stdout
	if cfg.Output != "-" {
		flag := os.O_CREATE|os.O_WRONLY
		if cfg.Append {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}
		if f, err := os.OpenFile(string(cfg.Output), flag, 0666); err != nil {
			fmt.Printf("Failed to open file %v. Cause: %v\n", cfg.Output, err)
			os.Exit(1)
		} else {
			file = f
			defer f.Close()
		}
	}
	for _, dir := range cfg.Dirs {
		err := ExportDirConst(file, dir)
		if err != nil {
			fmt.Printf("Failed to parse directory %v. Cause: %v\n", dir, err)
		}
	}
	if len(cfg.Args.Files) > 0 {
		err := ExportFilesConst(file, cfg.Package, cfg.Args.Files...)
		if err != nil {
			fmt.Printf("Failed to parse files. Cause: %v\n", err)
		}
	}
}
