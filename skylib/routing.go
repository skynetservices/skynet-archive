package skylib

import (
	"json"
	"log"
	"os"
)


// Function to retrieve a route by name from Doozer
// Returns a route, or error.
func GetRoute(name string) (r *Route, err os.Error) {

	data, _, err := DC.Get("/routes/"+name, nil)
	if err != nil {
		LogError(skynet.ERROR, err.String())
		return r, err
	}
	if len(data) > 0 {
		err := json.Unmarshal(data, &r)
		if err != nil {
			LogError(skynet.ERROR, err.String())
			return r, err
		}
		return r, nil
	}

	return r, os.NewError("No route found")

}
