package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const bindatafile = "bindata.go"

var (
	debug      bool
	bindataCmd string
)

func main() {
	flag.BoolVar(&debug, "debug", false, "")
	flag.StringVar(&bindataCmd, "go-bindata", "go-bindata", "command name or absolute path to the go-bindata binary")
	flag.Parse()

	log.SetFlags(0)

	if _, err := exec.LookPath(bindataCmd); err != nil {
		log.Println("Cannot find go-bindata executable in path")
		log.Println("Maybe you need: go get github.com/stephens2424/go-bindata-assetfs/...")
		os.Exit(1)
	}

	var args []string
	if debug {
		args = append(args, "debug")
	}

	args = append(args, flag.Args()...)

	cmd := exec.Command(bindataCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	in, err := os.Open(bindatafile)
	if err != nil {
		log.Println("Cannot read", bindatafile, err)
		return
	}
	out, err := os.Create("bindata_assetfs.go")
	if err != nil {
		log.Println("Cannot write 'bindata_assetfs.go'", err)
		return
	}
	r := bufio.NewReader(in)
	done := false
	for line, isPrefix, err := r.ReadLine(); err == nil; line, isPrefix, err = r.ReadLine() {
		if !isPrefix {
			line = append(line, '\n')
		}
		if _, err := out.Write(line); err != nil {
			log.Println("Cannot write to 'bindata_assetfs.go'", err)
			return
		}
		if !done && !isPrefix && bytes.HasPrefix(line, []byte("import (")) {
			if debug {
				fmt.Fprintln(out, "\t\"net/http\"")
			} else {
				fmt.Fprintln(out, "\t\"github.com/stephens2424/go-bindata-assetfs\"")
			}
			done = true
		}
	}
	if debug {
		fmt.Fprintln(out, `
func assetFS() http.FileSystem {
	for k := range _bintree.Children {
		return http.Dir(k)
	}
	panic("unreachable")
}`)
	} else {
		fmt.Fprintln(out, `
func assetFS() *assetfs.AssetFS {
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: k}
	}
	panic("unreachable")
}`)
	}
	// Close files BEFORE remove calls (don't use defer).
	in.Close()
	out.Close()
	if err := os.Remove(bindatafile); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot remove", bindatafile, err)
	}
}
