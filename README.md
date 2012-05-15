![logo](/bketelsen/skynet/raw/master/documentation/SkyNetLogo.png)

###SKYNET is under construction right now - docs don't match reality - please be patient for a few days

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

* Sending SIGUSR1 to a running process raises the log level one notch.
* Sending SIGUSR2 to a running process lowers the log level one notch.
* Sending SIGINT to a running process gracefully exits.

##Management
More management information to follow.  SKY is here!


##Communication
* Group name: Skynet-dev
* Group home page: http://groups.google.com/group/skynet-dev
* Group email address skynet-dev@googlegroups.com

##TODO:
Github Issues now the canonical source of issues for Skynet.

##Open Source - MIT Software License
Copyright (c) 2012 Brian Ketelsen

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
