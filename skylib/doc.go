/* 
   Copyright 2011 Brian Ketelsen


*/

/*
SkyNet is a framework for a distributed system of processes.

Each process in SkyNet receives its configuration from a centralized configuration repository (currently Doozer - possibly pluggable in the future).
Configuration changes are pushed to each process when changes to the processes running occur.  This means that starting a new service automatically
advertises that service's availability to the rest of the processes.

SkyNet is built on the premise that there will be at least three distinct process types:

Initiators - 
	Initiators are the source of inbound requests.  On a web-centric system, they'd be running HTTP listeners and accept web based requests.

Routers - 
	Routers are the "controller" of the system, they call services according to the stored route configuration that matches the request type.
	Technically routers are optional, but if they're not used, Initiators will call Services directly.

Services -
	Services are where the work gets done.  These are the processes that service the requests, process the API calls, get the external data, log the requests, authenticate the users, etc.

(Optional) Watchers -
	Watchers are tasks that run and know about the system, but aren't responding to individual requests.  An example of a watcher would be a process that watches the other processes in the system.


SkyNet uses Doozer to store configuration data about the available services and routes.  Configuration changes are pushed to Doozer, causing connected clients to immediately become aware of changed configurations.  

A typical transaction will come to an Initiator (via http for example) and be sent to a router that is providing the appropriate service to route that type of requests.  The Router checks its routes and calls the services
listed in its route configuration for that Route type.  Routes also define whether a service can be called Asynchronously (fire and forget) or whether the router must wait for a response.  For each service listed in the route
the Router calls the service passing in the request and response objects.  When all services are run, the router returns a response to the Initiator who is responsible for presenting the data to the remote client
appropriately.  In our HTTP example, this could mean translating to data using an HTML template, or an XML template.

TODO:
There are several things that can be improved in SkyNet.  The code needs significant refactoring.  Too much duplication exists.


*/
package skylib
