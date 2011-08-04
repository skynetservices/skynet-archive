//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/paulbellamy/mango"
	"github.com/bketelsen/skynet/skylib"
	"log"
	"os"
	"rpc"
	"template"
)

//const sName = "Initiator.Web"

const homeTemplate = `<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'><html xmlns='http://www.w3.org/1999/xhtml' xml:lang='en' lang='en'><head></head><body id='body'><form action='/new' operation='POST'><div>Your Input Value<input type='text' name='YourInputValue' value=''></input></div>	</form></body></html>`
const responseTemplate = `<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'><html xmlns='http://www.w3.org/1999/xhtml' xml:lang='en' lang='en'><head></head><body id='body'>{.repeated section resp.Errors} There were errors:<br/>{@}<br/>{.end}<div>Your Output Value: {resp.YourOutputValue}</div>	</body></html>	`

// Call the RPC service on the router to process the GetUserDataRequest.
func submitGetUserDataRequest(cr *skylib.SkynetRequest) (*skylib.SkynetResponse, os.Error) {
	var GetUserDataResponse *skylib.SkynetResponse

	sig := "RouteService"
	client, err := skylib.GetRandomClientBySignature(sig)
	if err != nil {
		if GetUserDataResponse == nil {
			GetUserDataResponse = &skylib.SkynetResponse{}
		}
		GetUserDataResponse.Errors = append(GetUserDataResponse.Errors, err.String())
		return GetUserDataResponse, err
	}
	err = client.Call(sig + ".RouteGetUserDataRequest", cr, &GetUserDataResponse)
	if err != nil {
		if GetUserDataResponse == nil {
			GetUserDataResponse = &skylib.SkynetResponse{}

		}
		GetUserDataResponse.Errors = append(GetUserDataResponse.Errors, err.String())
	}

	return GetUserDataResponse, nil
}

// Handler function to accept the submitted form post with the SSN
func submitHandler(env mango.Env) (mango.Status, mango.Headers, mango.Body) {

	log.Println("Submit GetUserData Request")
	inputs := make(map[string]interface{})
	inputs["YourInputValue"] = env.Request().FormValue("YourInputValue")
	cr := &skylib.SkynetRequest{Params: inputs}

	resp, err := submitGetUserDataRequest(cr)
	if err != nil {
		log.Println(err.String())
	}
	log.Println(resp)

	buffer := &bytes.Buffer{}
	respTmpl.Execute(buffer, map[string]interface{}{
		"resp": resp,
	})
	return 200, mango.Headers{}, mango.Body(buffer.String())
}


// Handler function to display the social form
func homeHandler(env mango.Env) (mango.Status, mango.Headers, mango.Body) {
	buffer := &bytes.Buffer{}
	homeTmpl.Execute(buffer, nil)
	return 200, mango.Headers{}, mango.Body(buffer.String())
}

var homeTmpl *template.Template
var respTmpl *template.Template

func main() {
	// Pull in command line options or defaults if none given
	flag.Parse()

	agent := skylib.NewAgent()
	agent.Start()

	homeTmpl = template.MustParse(homeTemplate, nil)
	respTmpl = template.MustParse(responseTemplate, nil)

	rpc.HandleHTTP()

	portString := fmt.Sprintf("%s:%d", *skylib.BindIP, *skylib.Port)

	stack := new(mango.Stack)
	stack.Address = portString

	routes := make(map[string]mango.App)
	routes["/"] = homeHandler
	routes["/new"] = submitHandler
	stack.Middleware(mango.Routing(routes))
	stack.Run(nil)
}
