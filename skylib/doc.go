//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

/*
SkyNet is a framework for a distributed system of processes.

Each process in SkyNet receives its configuration from a centralized configuration repository (currently Doozer - possibly pluggable in the future).
Configuration changes are pushed to each process when changes to the processes running occur.  This means that starting a new service automatically
advertises that service's availability to the rest of the processes.



Services -
	Services are where the work gets done.  These are the processes that service the requests, process the API calls, get the external data, log the requests, authenticate the users, etc.

(Optional) Watchers -
	Watchers are tasks that run and know about the system, but aren't responding to individual requests.  An example of a watcher would be a process that watches the other processes in the system.


SkyNet uses Doozer to store configuration data about the available services.  Configuration changes are pushed to Doozer, causing connected clients to immediately become aware of changed configurations.  

*/
package skylib
