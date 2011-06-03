#Skynet

##What is Skynet?
SkyNet is a framework for a distributed system of processes.  Skynet was designed for serving API requests across a cluster of servers, either on the "cloud" or in your datacenter(s).

##Why?
Skynet was designed with the assumption that services, processes and servers die, go away, become unreachable, crash, and generally don't work.  Search for the Chaos Monkey.  We wanted a system where each process was responsible for one thing, and any number of those processes could be started across any number of machines for recoverability, reliability and scalability.

##Shut up and tell me what to do!

	$ goinstall github.com/bketelsen/skynet/skygen
	$ goinstall github.com/bketelsen/skynet/skylib
	$ skygen -packageName=myCompany -serviceName=GetWidgets -targetFullPath="/Users/bketelsen/skynetTest/"

The skygen command generates a source tree with a running sample application.  After running skygen, cd into your target directory and build each service.  We use the awesome [go-gb](https://github.com/skelterjohn/go-gb).  Using gb, you simply issue the command "gb" from the root directory.  Each service will be compiled, and the executable will be named the same as its containing folder.  If you're following along, you'll have:

	skynetTest/
	skynetTest/bin/
	skynetTest/bin/router
	skynetTest/bin/service
	skynetTest/bin/watcher
	skynetTest/bin/reaper
	skynetTest/bin/initiator
			
Before you can run skynet you'll need to have at least one [doozer](https://github.com/ha/doozerd) process running.  
Now start each service, on a different port:

	$ bin/service -name=getwidgets -port=9200 &
	$ bin/initiator -name=webinitiator -port=9300 &
	$ bin/router -name=router -port=9100 &
 	$ bin/reaper -name=reaper -port=9000 &

If you don't specify a -logFileName parameter, they'll all default to using the same log file.  Now open a web browser and aim it at http://127.0.0.1:9300 
Enter something in the form and hit enter.  You should get a "Hello World" response.  

To really spice up your life, start up multiples of each process:

	$ bin/service -name=getwidgets2 -port=9201 &
	$ bin/initiator -name=webinitiator2 -port=9301 &
	$ bin/router -name=router2 -port=9101 &
	$ bin/reaper -name=reaper2 -port=9001 &
	
Connect to http://127.0.0.1:9300 or :9301 and see the same thing.  Kill the first router you started and submit a request... the second will handle the call.  Kill all of the services, skynet will return a pretty error message letting you know that there weren't any services available to handle the request.  

Now, go to http://127.0.0.1:9100/debug/vars (if you haven't killed that router process... if you have find any other process and substitute the port).  Skynet automatically exports statistical variables in JSON format:

	{
	...
	"RouteService.RouteGetACHDataRequest-goroutines": 29,
	"cmdline": ["bin/routers","-name=router","-port=9100"],
	"RouteService.RouteGetACHDataRequest-processed": 3030,
	"RouteService.RouteGetACHDataRequest-errors": 2
	}

##How?
Each process in SkyNet receives its configuration from a centralized configuration repository (currently Doozer - possibly pluggable in the future).  Configuration changes are pushed to each process when new skynet services are started.  This means that starting a new service automatically
advertises that service's availability to the rest of the members of the skynet cluster.

##What does it look like?
SkyNet is built on the premise that there will be at least three distinct process types:

1. Initiators - Initiators are the source of inbound requests.  On a web-centric system, they'd be running HTTP listeners and accept web based requests.  That isn't required, however.  We have initiators for flat files and TCP connections, too.  If you can get the bytes in using Go, it can be an initiator.
1. Routers - 	Routers are the "controller" of the system, they call services according to the stored route configuration that matches the request type.(Technically routers are optional, but if they're not used, Initiators will call Services directly.  This is an advanced configuration.)
1. Services -Services are where the work gets done.  These are the processes that service the requests, process the API calls, get the external data, log the requests, authenticate the users, etc.
1. (Optional) Watchers -Watchers are tasks that run and know about the system, but aren't responding to individual requests.  An example of a watcher would be a process that watches the other processes in the system and reports on statistics or availability.  The Reaper is a specialized watcher that checks each Skynet cluster member, culling dead processes from the configuration file.


##Dependencies
SkyNet uses Doozer to store configuration data about the available services and routes.  Configuration changes are pushed to Doozer, causing connected clients to immediately become aware of changed configurations.  

##Transaction Flow
A typical transaction will come to an Initiator (via http for example) and be sent to a router that is providing the appropriate service to route that type of requests.  The Router checks its routes and calls the services listed in its route configuration for that Route type.  Routes also define whether a service can be called Asynchronously (fire and forget) or whether the router must wait for a response.  For each service listed in the route the Router calls the service passing in the request and response objects.  When all services are run, the router returns a response to the Initiator who is responsible for presenting the data to the remote client appropriately.  In our HTTP example, this could mean translating to data using an HTML template, or an XML/JSON template.

##TODO:
* Write a watcher that consolidates all of the json/expvars and puts them in a pretty graph/chart/widget that makes managers and sysadmins happy
* The code is just plain ugly.  It needs clean up in every corner.  It is an extraction of a work in progress.
* Routes should be viewable and editable using a pretty web interface
* Pluggable configuration - Redis?
* Support JSON-RPC as a transport instead of Go's native RPC.  This would allow skynet cluster members written in other languages.
* Cache or pool RPC connections between cluster members.
* Examples
* Video demo