//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

const initiatorTemplate = `package main


import "github.com/bketelsen/skynet/skylib"
import "<%PackageName%>"
import "log"
import "os"
import "http"
import "template"
import "flag"
import "fmt"
import "rpc"

const sName = "Initiator.Web"

const homeTemplate = ` + "`<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'><html xmlns='http://www.w3.org/1999/xhtml' xml:lang='en' lang='en'><head></head><body id='body'><form action='/new' method='POST'><div>Your Input Value<input type='text' name='YourInputValue' value=''></input></div>	</form></body></html>`" + `
const responseTemplate = ` + "`<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'><html xmlns='http://www.w3.org/1999/xhtml' xml:lang='en' lang='en'><head></head><body id='body'>{.repeated section resp.Errors} There were errors:<br/>{@}<br/>{.end}<div>Your Output Value: {resp.YourOutputValue}</div>	</body></html>	`" + `


// Call the RPC service on the router to process the <%ServiceName%>Request.
func submit<%ServiceName%>Request(cr *<%PackageName%>.<%ServiceName%>Request) (*<%PackageName%>.<%ServiceName%>Response, os.Error) {
	var <%ServiceName%>Response *<%PackageName%>.<%ServiceName%>Response

	client, err := skylib.GetRandomClientByProvides("RouteService.Route<%ServiceName%>Request")
	if err != nil {
		if <%ServiceName%>Response == nil {
			<%ServiceName%>Response = &<%PackageName%>.<%ServiceName%>Response{}
		}
		<%ServiceName%>Response.Errors = append(<%ServiceName%>Response.Errors, err.String())
		return <%ServiceName%>Response, err
	}
	err = client.Call("RouteService.Route<%ServiceName%>Request", cr, &<%ServiceName%>Response)
	if err != nil {
		if <%ServiceName%>Response == nil {
			<%ServiceName%>Response = &<%PackageName%>.<%ServiceName%>Response{}

		}
		<%ServiceName%>Response.Errors = append(<%ServiceName%>Response.Errors, err.String())
	}

	return <%ServiceName%>Response, nil
}

// Handler function to accept the submitted form post with the SSN
func submitHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Submit <%ServiceName%> Request")
	cr := &<%PackageName%>.<%ServiceName%>Request{YourInputValue: r.FormValue("YourInputValue")}

	resp, err := submit<%ServiceName%>Request(cr)
	if err != nil {
		skylib.LogError(err.String())
	}
	
	go skylib.LogDebug(resp)

	respTmpl.Execute(w, map[string]interface{}{
		"resp": resp,
	})
}


// Handler function to display the social form
func homeHandler(w http.ResponseWriter, r *http.Request) {
	homeTmpl.Execute(w, nil)

}

var homeTmpl *template.Template
var respTmpl *template.Template

func main() {

	var err os.Error

	// Pull in command line options or defaults if none given
	flag.Parse()

	f, err := os.OpenFile(*skylib.LogFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}

	skylib.Setup(sName)

	homeTmpl = template.MustParse(homeTemplate, nil)
	respTmpl = template.MustParse(responseTemplate, nil)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/new", submitHandler)

	rpc.HandleHTTP()

	portString := fmt.Sprintf("%s:%d", *skylib.BindIP, *skylib.Port)

	err = http.ListenAndServe(portString, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
}
`
