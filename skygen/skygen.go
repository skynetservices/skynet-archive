//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"flag"
	"fmt"
	"os"
	"template"
)


// Service generation flags
var excludeInitiator *bool = flag.Bool("excludeInitiator", false, "don't generate an initiator")
var excludeRouter *bool = flag.Bool("excludeRouter", false, "don't generate a router")
var excludeService *bool = flag.Bool("excludeService", false, "don't generate a service")
var excludeWatcher *bool = flag.Bool("excludeWatcher", false, "don't generate a watcher")
var excludeReaper *bool = flag.Bool("excludeReaper", false, "don't generate a reaper")

// Configuration flags
var PackageName *string = flag.String("packageName", "myCompany", "namespace of the Go package to generate")
var ServiceName *string = flag.String("serviceName", "GetUsers", "API function to be provided by SkyNet")
var TargetFullPath *string = flag.String("targetFullPath", "./myskynet", "Full path of target for skynet generation; best if not in skynet working tree.")

func printToDo() {
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Println(f.Name, f.Value)
	})
}

func generateLibrary() {
	//mkdir for the packagename
	err := os.MkdirAll(*TargetFullPath+"/"+*PackageName, 0755)
	if err != nil {
		fmt.Println("Unable to create directory, ", PackageName, err.String())
	}
	f, err := os.Create(*TargetFullPath + "/" + *PackageName + "/package.go")
	if err != nil {
		fmt.Println(err.String())
	}
	defer f.Close()
	var templ *template.Template
	templ = template.New(nil)
	templ.SetDelims("<%", "%>")
	err = templ.Parse(packageTemplate)
	if err != nil {
		fmt.Println(err.String())
	}
	err = templ.Execute(f, map[string]interface{}{
		"PackageName": *PackageName,
		"ServiceName": *ServiceName,
	})

	if err != nil {
		fmt.Println(err.String())
	}
}

func generateInitiator() {
	//mkdir for the initiator
	err := os.MkdirAll(*TargetFullPath+"/initiators/web/", 0755)
	if err != nil {
		fmt.Println("Unable to create directory, ", err.String())
	}
	f, err := os.Create(*TargetFullPath + "/initiators/web/" + "web.go")
	if err != nil {
		fmt.Println(err.String())
	}
	defer f.Close()
	var templ *template.Template
	templ = template.New(nil)
	templ.SetDelims("<%", "%>")
	err = templ.Parse(initiatorTemplate)
	if err != nil {
		fmt.Println(err.String())
	}
	err = templ.Execute(f, map[string]interface{}{
		"PackageName": *PackageName,
		"ServiceName": *ServiceName,
	})

	if err != nil {
		fmt.Println(err.String())
	}
}

func generateRouter() {
	//mkdir for the initiator
	err := os.MkdirAll(*TargetFullPath+"/router/", 0755)
	if err != nil {
		fmt.Println("Unable to create directory, ", err.String())
	}
	f, err := os.Create(*TargetFullPath + "/router/" + "router.go")
	if err != nil {
		fmt.Println(err.String())
	}
	defer f.Close()
	var templ *template.Template
	templ = template.New(nil)
	templ.SetDelims("<%", "%>")
	err = templ.Parse(routerTemplate)
	if err != nil {
		fmt.Println(err.String())
	}
	err = templ.Execute(f, map[string]interface{}{
		"PackageName": *PackageName,
		"ServiceName": *ServiceName,
	})

	if err != nil {
		fmt.Println(err.String())
	}
}

func generateService() {
	//mkdir for the initiator
	err := os.MkdirAll(*TargetFullPath+"/service/", 0755)
	if err != nil {
		fmt.Println("Unable to create directory, ", err.String())
	}
	f, err := os.Create(*TargetFullPath + "/service/" + "service.go")
	if err != nil {
		fmt.Println(err.String())
	}
	defer f.Close()
	var templ *template.Template
	templ = template.New(nil)
	templ.SetDelims("<%", "%>")
	err = templ.Parse(serviceTemplate)
	if err != nil {
		fmt.Println(err.String())
	}
	err = templ.Execute(f, map[string]interface{}{
		"PackageName": *PackageName,
		"ServiceName": *ServiceName,
	})

	if err != nil {
		fmt.Println(err.String())
	}
}

func generateReaper() {
	//mkdir for the initiator
	err := os.MkdirAll(*TargetFullPath+"/watchers/reaper/", 0755)
	if err != nil {
		fmt.Println("Unable to create directory, ", err.String())
	}
	f, err := os.Create(*TargetFullPath + "/watchers/reaper/" + "reaper.go")
	if err != nil {
		fmt.Println(err.String())
	}
	defer f.Close()
	var templ *template.Template
	templ = template.New(nil)
	templ.SetDelims("<%", "%>")
	err = templ.Parse(reaperTemplate)
	if err != nil {
		fmt.Println(err.String())
	}
	err = templ.Execute(f, map[string]interface{}{
		"PackageName": *PackageName,
		"ServiceName": *ServiceName,
	})

	if err != nil {
		fmt.Println(err.String())
	}
}
func generateWatcher() {
	//mkdir for the initiator
	err := os.MkdirAll(*TargetFullPath+"/watchers/generic/", 0755)
	if err != nil {
		fmt.Println("Unable to create directory, ", err.String())
	}
	f, err := os.Create(*TargetFullPath + "/watchers/generic/" + "generic.go")
	if err != nil {
		fmt.Println(err.String())
	}
	defer f.Close()
	var templ *template.Template
	templ = template.New(nil)
	templ.SetDelims("<%", "%>")
	err = templ.Parse(watcherTemplate)
	if err != nil {
		fmt.Println(err.String())
	}
	err = templ.Execute(f, map[string]interface{}{
		"PackageName": *PackageName,
		"ServiceName": *ServiceName,
	})

	if err != nil {
		fmt.Println(err.String())
	}
}
//Skynet Generator creates the files necessary for a skynet installation
// Minimum required are an Initiator and a Service
//
//
// flags: -excludeInitiator
//		  -excludeRouter
//        -excludeService
//        -excludeWatcher
//        -excludeReaper
//        -packageName=myPackage
//        -serviceName=GetUsers
// 
func main() {
	flag.Parse()

	printToDo()

	generateLibrary()
	
	if !*excludeInitiator {
		generateInitiator()
	}
	if !*excludeRouter {
		generateRouter()
	}
	if !*excludeReaper {
		generateReaper()
	}
	if !*excludeService {
		generateService()
	}
	if !*excludeWatcher {
		generateWatcher()
	}
}
