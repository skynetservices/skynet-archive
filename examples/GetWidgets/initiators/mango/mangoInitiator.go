package main

import (
	"github.com/paulbellamy/mango"
)

func Hello(env mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, mango.Headers{}, mango.Body("Hello World!")
}

func main() {
	stack := new(mango.Stack)
	stack.Address = ":3000"
	stack.Run(Hello)
}
