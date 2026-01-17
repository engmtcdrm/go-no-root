package internal

import (
	"fmt"

	"github.com/engmtcdrm/go-eggy"
)

var AllExamples = []eggy.Example{
	{Name: "Hello World Example", Fn: helloWorldExample},
	{Name: "Hi Mom Example", Fn: hiMomExample},
}

func helloWorldExample() {
	fmt.Println("Hello World!")
}

func hiMomExample() {
	fmt.Println("Hi Mom!")
}
