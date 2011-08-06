//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package skylib

import (
	"log"
	"json"
	"os"
	"fmt"
	"rand"
	"rpc"
	"rpc/jsonrpc"
	"strings"
)


var NOS *RegisteredNetworkServers

// Return a list of all RpcServices which provide the named Service.
func GetAllServiceProviders(classname string) (serverList []*RpcService) {
	for _, v := range NOS.Services {
		if v != nil && v.Provides == classname {
			serverList = append(serverList, v)
		}
	}
	return
}

func GetAllClientsByService(classname string) (clientList []*rpc.Client) {
	var newClient *rpc.Client
	var err os.Error
	serviceList := GetAllServiceProviders(classname)

	for i, s := range serviceList {
		hostString := fmt.Sprintf("%s:%d", s.IPAddress, s.Port)
		protocol := strings.ToLower(s.Protocol) // to be safe
		switch protocol {
		default:
			newClient, err = rpc.DialHTTP("tcp", hostString)
		case "json":
			newClient, err = jsonrpc.Dial("tcp", hostString)
		}

		if err != nil {
			LogWarn(fmt.Sprintf("Found %d nodes to provide service %s requested on %s, but failed to connect to #%d.",
				len(serviceList), classname, hostString, i))
			//NewError(NO_CLIENT_PROVIDES_SERVICE, classname)
			continue
		}
		clientList = append(clientList, newClient)
	}
	return
}

// This is simple today - it returns the first listed service that matches the request
// Load balancing needs to be applied here somewhere.
func GetRandomClientByService(classname string) (*rpc.Client, os.Error) {
	var newClient *rpc.Client
	var err os.Error
	serviceList := GetAllServiceProviders(classname)

	if len(serviceList) > 0 {
		chosen := rand.Int() % len(serviceList)
		s := serviceList[chosen]

		hostString := fmt.Sprintf("%s:%d", s.IPAddress, s.Port)
		protocol := strings.ToLower(s.Protocol) // to be safe
		switch protocol {
		default:
			newClient, err = rpc.DialHTTP("tcp", hostString)
		case "json":
			newClient, err = jsonrpc.Dial("tcp", hostString)
		}

		if err != nil {
			LogWarn(fmt.Sprintf("Found %d nodes to provide service %s requested on %s, but failed to connect.",
				len(serviceList), classname, hostString))
			return nil, NewError(NO_CLIENT_PROVIDES_SERVICE, classname)
		}

	} else {
		LogWarn(fmt.Sprintf("Found no node to provide service %s.", classname))
		return nil, NewError(NO_CLIENT_PROVIDES_SERVICE, classname)
	}
	return newClient, nil
}


// on startup load the configuration file. 
// After the config file is loaded, we set the global config file variable to the
// unmarshaled data, making it useable for all other processes in this app.
func LoadRegistry() {
	data, _, err := DC.Get("/servers/config/networkservers.conf", nil)
	if err != nil {
		log.Panic(err.String())
	}
	if len(data) > 0 {
		setRegistry(data)
		return
	}
	LogError("Error loading default config - no data found")
	NOS = &RegisteredNetworkServers{}
}

func RemoveService(i int) {

	newServices := make([]*RpcService, 0)

	for k, v := range NOS.Services {
		if k != i {
			if v != nil {
				newServices = append(newServices, v)
			}
		}
	}
	NOS.Services = newServices
	b, err := json.Marshal(NOS)
	if err != nil {
		log.Panic(err.String())
	}
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.String())
	}
	_, err = DC.Set("/servers/config/networkservers.conf", rev, b)
	if err != nil {
		log.Panic(err.String())
	}

}

func RemoveFromRegistry(r *RpcService) {

	newServices := make([]*RpcService, 0)

	for _, v := range NOS.Services {
		if v != nil {
			if !v.Equal(r) {
				newServices = append(newServices, v)
			}

		}
	}
	NOS.Services = newServices
	b, err := json.Marshal(NOS)
	if err != nil {
		log.Panic(err.String())
	}
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.String())
	}
	_, err = DC.Set("/servers/config/networkservers.conf", rev, b)
	if err != nil {
		log.Panic(err.String())
	}
}

func AddToRegistry(r *RpcService) {
	for _, v := range NOS.Services {
		if v != nil {
			if v.Equal(r) {
				LogInfo(fmt.Sprintf("Skipping adding %s : alreday exists.", v.Provides))
				return // it's there so we don't need an update
			}
		}
	}
	NOS.Services = append(NOS.Services, r)
	LogDebug("Added", r.Provides, r.Protocol)
	b, err := json.Marshal(NOS)
	if err != nil {
		log.Panic(err.String())
	}
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.String())
	}
	_, err = DC.Set("/servers/config/networkservers.conf", rev, b)
	if err != nil {
		log.Panic(err.String())
	}
}

// unmarshal data from remote store into global config variable
func setRegistry(data []byte) {
	err := json.Unmarshal(data, &NOS)
	if err != nil {
		log.Panic(err.String())
	}
}

// Watch for remote changes to the config file.  When new changes occur
// reload our copy of the config file.
// Meant to be run as a goroutine continuously.
func WatchRegistry() {
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.String())
	}

	for {
		// blocking wait call returns on a change
		ev, err := DC.Wait("/servers/config/networkservers.conf", rev)
		if err != nil {
			log.Panic("Error waiting on config: " + err.String())
		}
		log.Println("Received new configuration.  Setting local config.")
		setRegistry(ev.Body)

		rev = ev.Rev + 1
	}

}
