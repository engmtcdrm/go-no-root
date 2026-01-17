package main

import (
	"fmt"

	"github.com/engmtcdrm/go-eggy"
	noroot "github.com/engmtcdrm/go-no-root"
	"github.com/engmtcdrm/go-no-root/examples/internal"
)

func main() {
	eggy.NewExamplePrompt(internal.AllExamples).
		Title("My Package Examples").
		Show()

	ps, err := noroot.GetCurrentGoProcess()
	if err != nil {
		panic(err)
	}
	fmt.Println("Process PID:", ps.Pid())
}
