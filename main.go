package main

import (
	"debug/dwarf"
	"fmt"
	"github.com/go-delve/delve/pkg/dwarf/op"
	"github.com/go-delve/delve/pkg/dwarf/reader"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/kr/pretty"
	"github.com/spf13/pflag"
	"log"
	"math/rand"
)

var (
	filename string
)

func InitFlags() {
	pflag.StringVar(&filename, "binary", "binary", "binary")
	pflag.Parse()
}

func init() {
	InitFlags()
}

// go:noinline
func main() {
	bi := proc.NewBinaryInfo("linux", "amd64")
	if err := bi.LoadBinaryInfo(filename, 0 /* entryPoint */, nil /* debugInfoDirs */); err != nil {
		panic(err)
	}
	img := bi.Images[0]
	//img.GetDwarfTree()
	f, ok := bi.LookupFunc["main.xxx"]
	if !ok {
		panic("!!!")
	}

	funcDwarfTree, err := img.GetDwarfTree(f.Offset())
	if err != nil {
		panic(err)
	}

	pc := func2PC(f)
	fmt.Printf("pc: %d\n", pc)
	_, line, f2 := bi.PCToLine(pc)
	fmt.Printf("line: %d, f2: %v\n", line, f2)
	if f2.Name != f.Name {
		log.Printf("instrumenting different function: %s", f2.Name)
	}
	varEntries := reader.Variables(funcDwarfTree, f.Entry, line, 0 /* flags */)
	for i, v := range varEntries {
		log.Printf("looking up variable: %s", pretty.Sprint(v))
		addr, pieces, descr, err := bi.Location(v.Entry, dwarf.AttrLocation, pc, op.DwarfRegisters{})
		if err != nil {
			panic(err)
		}
		log.Printf("var %d: addr: %x, pieces: %v, descr: %v", i, addr, pieces, descr)
	}

	fmt.Println(xxx(10))
}

func func2PC(fn *proc.Function) uint64 {
	if fn.Entry != 0 {
		return fn.Entry
	}
	if len(fn.InlinedCalls) != 0 {
		return fn.InlinedCalls[0].LowPC
	}
	return 0
}

// go:noinline
func xxx(a int) int {
	return rand.Intn(10) + a
}
