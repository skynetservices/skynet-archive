package main


import "github.com/bketelsen/skynet/skylib"
import "myStartup"
import "log"
import "os"
import "http"
import "template"
import "flag"
import "fmt"
import "rpc"

const sName = "Initiator.Web"

const homeTemplate = `<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'><html xmlns='http://www.w3.org/1999/xhtml' xml:lang='en' lang='en'><head></head><body id='body'><form action='/new' method='POST'><div>Your Input Value<input type='text' name='YourInputValue' value=''></input></div>	</form></body></html>`
const responseTemplate = `<!DOCTYPE html PUBLIC '-//W3C//DTD XHTML 1.0 Transitional//EN' 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd'><html xmlns='http://www.w3.org/1999/xhtml' xml:lang='en' lang='en'><head></head><body id='body'>{.repeated section resp.Errors} There were errors:<br/>{@}<br/>{.end}<div>Your Output Value: {resp.YourOutputValue}</div>	</body></html>	`


// Call the RPC service on the router to process the GetUserDataRequest.
func submitGetUserDataRequest(cr *myStartup.GetUserDataRequest) (*myStartup.GetUserDataResponse, os.Error) {
	var GetUserDataResponse *myStartup.GetUserDataResponse

	client, err := skylib.GetRandomClientByProvides("RouteService.RouteGetUserDataRequest")
	if err != nil {
		if GetUserDataResponse == nil {
			GetUserDataResponse = &myStartup.GetUserDataResponse{}
		}
		GetUserDataResponse.Errors = append(GetUserDataResponse.Errors, err.String())
		return GetUserDataResponse, err
	}
	err = client.Call("RouteService.RouteGetUserDataRequest", cr, &GetUserDataResponse)
	if err != nil {
		if GetUserDataResponse == nil {
			GetUserDataResponse = &myStartup.GetUserDataResponse{}

		}
		GetUserDataResponse.Errors = append(GetUserDataResponse.Errors, err.String())
	}

	return GetUserDataResponse, nil
}

// Handler function to accept the submitted form post with the SSN
func submitHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Submit GetUserData Request")
	cr := &myStartup.GetUserDataRequest{YourInputValue: r.FormValue("YourInputValue")}

	resp, err := submitGetUserDataRequest(cr)
	if err != nil {
		log.Println(err.String())
	}
	log.Println(resp)

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
