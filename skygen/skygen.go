//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"flag"
	"fmt"
	// "os"
	// "template"
)

// Configuration flags
var PackageName *string = flag.String("packageName", "myCompany", "namespace of the Go package to generate")
var ServiceName *string = flag.String("serviceName", "GetUsers", "API function to be provided by SkyNet")
var TargetFullPath *string = flag.String("targetFullPath", "./myskynet", "Full path of target for skynet generation; best if not in skynet working tree.")

func printToDo() {
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Println(f.Name, f.Value)
	})
}

func generateService() {
	//TODO
}

//Skynet Generator creates the files necessary for a skynet installation
// Minimum required are an Initiator and a Service
//
//
// flags: 
//        -packageName=myPackage
//        -serviceName=GetUsers
// 
func main() {
	flag.Parse()

	printToDo()
	generateService()

}
