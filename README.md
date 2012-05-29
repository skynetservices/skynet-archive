![logo](/bketelsen/skynet/raw/master/documentation/SkyNetLogo.png)

##Introduction
Skynet is a system for building massively distributed apps in Go.

##Tell me more:
Servers die, stop communicating, catch on fire, get killed by robots from the future, and should not be trusted. If your site won’t work with a Chaos Monkey, it isn’t safe. Enter Skynet. Each Skynet module is self–contained, self–aware, and self–replicating – if you have one server with an authentication module on it, and that server melts, Skynet will notice, kill it, and automatically create a new one. (if you let it)

Skynet probably won’t die unless your data center gets hit by a comet.  We recommend at least 2 data centers in that scenario.

Skynet Services are where the work gets done.  These are the processes that service the requests, process the API calls, get the external data, log the requests, authenticate the users, etc. 

			
Before you can run skynet you'll need to have at least one [doozer](https://github.com/ha/doozerd) process running.  

##How?
Each process in SkyNet receives its configuration from a centralized configuration repository (currently Doozer - possibly pluggable in the future).  Configuration changes are pushed to each process when new skynet services are started.  This means that starting a new service automatically
advertises that service's availability to the rest of the members of the skynet cluster.

SkyNet uses Doozer to store configuration data about the available services.  Configuration changes are pushed to Doozer, causing connected clients to immediately become aware of changed configurations.  

##Running Processes
* Sending SIGINT to a running process gracefully exits.

######Work in Progress
* Sending SIGUSR1 to a running process raises the log level one notch.
* Sending SIGUSR2 to a running process lowers the log level one notch.

## Doozer
Skynet makes heavy usage of Doozer. Both clients and services will take a DoozerConfig so that it knows how to communicate with doozer. In the examples directory there is a shell script to startup a cluster of doozer instances locally for testing.

<pre>
type DoozerConfig struct {
	Uri          string
	BootUri      string
	AutoDiscover bool
}
</pre>

* Uri - ip/port of your doozer instance, this can be a comma separated list as well (doozer:8046, 127.0.0.1:8046)
* BootUri - If you are using DzNs this is the ip/port of an instance of your boot cluster (doozer:8046, 127.0.0.1:8046)
* AutoDiscover - true/false, Should this service or client discover other doozer instances in the doozer cluster, and use them for failback in case it looses it's current connection, as well as monitor any addition/removal from the doozer cluster

##Services
Services are the heart of your skynet clusters, they will accept requests via msgpack rpc requests. Keep in mind that a Service may also be a client. In the case of a Composite style application, a request could be made to one service that makes requests either synchronously or asynchronously to additional skynet services.

#####Sweet! How do I create a service?
Provided you have a doozer instance setup. It's pretty simple. Just create your service, with any methods you want exposed via rpc, and make sure it implements skylib.ServiceInterface

<pre>
type ServiceInterface interface {
	Started(s *Service)
	Stopped(s *Service)
	Registered(s *Service)
	Unregistered(s *Service)
}
</pre>

Then call skylib.CreateService() passing it, a ServiceConfig, and and pointer to your custom service. Then call:

<pre>
service.Start(true)
</pre>

The boolean flag specifies whether the service should immediately register itself with the cluster.

<pre>
type BindAddr struct {
	IPAddress string
	Port      int
}

type ServiceConfig struct {
	Log         *log.Logger `json:"-"`
	Name        string
	Version     string
	Region      string
	ServiceAddr *BindAddr
	AdminAddr   *BindAddr
	DoozerConfig *DoozerConfig `json:"-"`
}
</pre>

Checkout the examples/service directory for a full example, also a call to skylib.GetServiceConfigFromFlags() will, allow you to get all the config params from flags passed in via the command line.

##Clients
Clients are just as simple. They start with a ClientConfig:

<pre>
type ClientConfig struct {
	Log         *log.Logger `json:"-"`
	DoozerConfig *DoozerConfig `json:"-"`
}
</pre>

Then a call to:

<pre>
skylib.GetService(name string, version string, region string, host string) (*ServiceClient)
</pre>

* name - the name of the service you want to connect to, this is specified in your ServiceConfig
* version - the version of the service, in case you have multiple versions up and running for backward compatibility, or you are in the middle of an upgrade / deploy
* region - If you only want a connection(s) to instances of this service in a specific region (specified by the service) in case you want to keep a particular request in data center etc.
* host - Similar to region, this restricts connections to only the specified host, an example might be if you only want to connect to the host the current client is on.

Any empty values "", are considered to mean any/all instances matching the other supplied criteria.

This call returns a pointer to a ServiceClient, think of this as a connection pool, to instances of the service you requested, matching your criteria. It will always point contain connections to live instances, and readjust itself as the cluster changes, and recover from connection failures.

From here just call your RPC method:

<pre>
serviceClient.Send("echo", "I'm connected!!")
</pre>

Checkout the examples/client. directory for a full example.

##Management
The "sky" command is your management gateway into the skynet cluster. It will allow you to probe the network and look for services/versions, hosts, regions etc in your cluster. As well as run administration commands to operate on instances that match the criteria (*admin commands are on the way, search functionality is here)

<pre>
Usage:
	 sky -option1=value -option2=value command <arguments>

Commands:

	hosts: List all hosts available that meet the specified criteria
		-service - limit results to hosts running the specified service
		-version - limit results to hosts running the specified version of the service (-service required)
		-region - limit results to hosts in the specified region
	instances: List all instances available that meet the specified criteria
		-service - limit results to instances of the specified service
		-version - limit results to instances of the specified version of service
		-region - limit results to instances in the specified region
		-host - limit results to instances on the specified host
	regions: List all regions available that meet the specified criteria
	services: List all services available that meet the specified criteria
		-host - limit results to the specified host
		-region - limit results to hosts in the specified region

	service-versions: List all services available that meet the specified criteria
		-service - service name (required)
		-host - limit results to the specified host
		-region - limit results to hosts in the specified region

	topology: Print detailed heirarchy of regions/hosts/services/versions/instances
		-service - limit results to instances of the specified service
		-version - limit results to instances of the specified version of service
		-region - limit results to instances in the specified region
		-host - limit results to instances on the specified host
</pre>

##Internals
#####Query
The sky command and the client connectivity logic is all backed by skylib.Query. A struct that can be used to search the cluster for instances of Services matching specified criteria. It's exposed for any custom need you may have for searching the cluster.

When the  cpu/memory/load monitoring is implemented Query will also be expanded to support searching based on this criteria.

<pre>
type Query struct {
	Service    string
	Version    string
	Host       string
	Region     string
	DoozerConn *DoozerConnection
}
</pre>

The only required field here is a pointer to a doozer connection. All other fields are optional, any field not supplied will be considered as any/all.

From here you can use any of the following

<pre>
query.FindInstances()
</pre>

Which will return a pointer to an array of Service pointers

<pre>
// *[]*Service

// Refer to above for ServiceConfig structure
type Service struct {
	Config     *ServiceConfig
	Registered bool              `json:"-"`
}
</pre>

If you feel like checking out the source some other things Query allow you to do:
<pre>
query.FindHosts()
query.FindServices()
query.FindRegions()
query.ServiceVersions()
</pre>

## Work In Progress
#####Smart Connection Pools
ServiceClients's will have a pool of connections to a given service, to load balance across. Instances are already removed from skynet when they crash, but local pools will be smart enough to remove any connections to any instances that are no longer available and replace them with connections to valid instances to maintain pool size.

#####Process Monitoring / Restarting
Services will restart themselves a specified number of times after crashing and add themselves back to the pool.

#####Monitor CPU/Memory/Load
Skynet will be aware of system utilization and number of requests going to specific instances of a service, and will be able to have a configurable threshold for it to restart or remove itself from the pool of instances for it's particular Service/Version, and return itself to the queue when the system has leveled out, or restart has completed

#####Administration through sky command
You will be able to register/unregister instances from skynet, stop, restart instances in your skynet cluster just by using the "sky" command with flags to filter instances (refer to the "sky" command section for more details on how these filters work"

#####Time Series Data / Metrics
We all love metrics and graphs, skynet will make sure you get your daily fix. More than likely we will utilize something like graphite to log time series data regarding number of requests, which calls are being made, response times, cpu/memory/load so that you can determine the state of your skynet cluster quickly with a dashboard of metrics.

After this functionality has been added we'd like to create a nice web interface to be hit and see live statistics going on across the system at that moment and refreshed live.

#####Test Suite / Benchmarks / Refactoring / Docs
Skynet has evolved quite a bit from the original idea/implementation, much experimentation and R&D has been done to come up with the best approach, now that a lot of this has been finalized a full test suite will be coming, as well as some cleanup in areas of the codebase that were just quick prototypes to prove theories and need to be clean interfaces.

Also the addition of godocs, and other wiki pages to help better describe internals, and tips/tricks.

#####Skylib for other languages
Skylib is the core of skynet's internals and is how services/clients find each other, by implementing skylib in a variety of languages we will allow services and clients of many different languages to become a part of the skynet cluster.

##Communication
* Group name: Skynet-dev
* Group home page: http://groups.google.com/group/skynet-dev
* Group email address skynet-dev@googlegroups.com

##Issues:
Github Issues now the canonical source of issues for Skynet.

##Open Source - MIT Software License
Copyright (c) 2012 Brian Ketelsen

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
