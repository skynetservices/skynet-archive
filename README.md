![logo](/bketelsen/skynet/raw/master/documentation/SkyNetLogo.png)

##Introduction
Skynet is a communication protocol for building massively distributed apps in Go.
It is not constrained to Go, so it will lend itself nicely to polyglot environments.
The first planned language addition is Ruby.

##Tell me more:
Servers die, stop communicating, catch on fire, get killed by robots from the future, and should not be trusted.

If your site won’t work with a Chaos Monkey, it isn’t safe.
Enter Skynet. Each Skynet module is self–contained and self–aware – if you have one server with an authentication module on it, and that server melts, Skynet will notice, kill it, and automatically create a new one. (if you let it)

Skynet probably won’t die unless your data center gets hit by a comet.  We recommend at least 2 data centers in that scenario.

[Skynet Services](skynet/wiki/Services) are where the work gets done.  These are the processes that service the requests, process the API calls, get the external data, log the requests, authenticate the users, etc.


Before you can run skynet you'll need to have at least one [doozerd](skynet/wiki/Setting-up-a-Doozer-cluster) process running.

##How?
Each process in SkyNet receives its configuration from a centralized configuration repository (currently [Doozer](skynet/wiki/Setting-up-a-Doozer-cluster) - possibly pluggable in the future).

Configuration changes are pushed to each process when new skynet services are started.
This means that starting a new service automatically advertises that service's availability to the rest of the members of the skynet cluster.

Processes are monitored, and restarted when they die, and are removed from the cluster configuration management system so that clients do not create new connections. Currently connected clients will notice these removals and adjust their internal pool of connections as services are added/removed/die.

[https://github.com/bketelsen/skynet/wiki/Daemon](https://github.com/bketelsen/skynet/wiki/Daemon)

## Doozer
Skynet makes heavy usage of [Doozer](skynet/wiki/Setting-up-a-Doozer-cluster). You'll need at least 1 Doozer instance to run Skynet, but we recommend a cluster of multiple Doozer nodes to ensure high availability. With only 1 instance you leave yourself with a single point of failure.

##Services
[Services](skynet/wiki/Services) are the heart of your Skynet clusters, they will accept requests via bson rpc requests, although this is abstracted away, you won't have to deal with the protocol, you will just pass objects. Keep in mind that a Service may also be a client. In the case of a Composite style application, a request could be made to one service that makes requests either synchronously or asynchronously to additional Skynet services.

#####Sweet! How do I create a service?
Check out the service documentation page on the wiki: [https://github.com/bketelsen/skynet/wiki/Services](https://github.com/bketelsen/skynet/wiki/Services)

Examples can be found in the *examples/* directory.

##Clients
[Clients](skynet/wiki/Clients) are responsible for sending requests to Skynet services, and processing their requests.

Clients contain a pool of connections to a given service, up to a specified size to load balance requests across. Instances are removed from Skynet when they crash, the pools are smart enough to remove any connections to any instances that are no longer available and replace them with connections to valid instances to maintain the pool.

To use the Go client in your own program to call Skynet services, see [Creating a simple client](https://github.com/bketelsen/skynet/wiki/Client-Tutorial)
The Go client is part of this project. Clients in other languages are also available:
* [Ruby Skynet Client](http://github.com/ClarityServices/ruby_skynet)
* [PHP Skynet Client](http://github.com/mikespook/php_skynet)

To build a Client in your favorite language see the [Skynet Protocol Guide](https://github.com/bketelsen/skynet/blob/master/protocol.md)

##Management

####Sky
The "[sky](skynet/wiki/Sky)" command is a management gateway into the Skynet cluster. It will allow you to probe the network and look for services/versions, hosts, regions etc in your cluster, as well as run administration commands to operate on instances that match the criteria. You can register/unregister/stop/restart services as well as deploy new services to hosts matching your filters.

####Interactive Shell
Another option is to use the interactive shell "<b>sky cli</b>". Which will open a shell you can interact with. Setting filters will allow any future commands to only apply to resources that meet those conditions. It supports history, and tab completion of commands, as well as services, hosts, regions, versions Skynet is already aware of.

[https://github.com/bketelsen/skynet/wiki/Sky](https://github.com/bketelsen/skynet/wiki/Sky)

####Dashboard
The [dashboard](skynet/wiki/Dashboard) is a live updating web ui. Where you can see the current topology of your network, what regions/hosts/instances are up, average response times, last request, number of connections, if they are registered or not.

In the future you will be able to live search your logs, as well as see graph data surrounding the health of your cluster.

![picture](/bketelsen/skynet/raw/master/documentation/dashboard.png)

[https://github.com/bketelsen/skynet/wiki/Dashboard](https://github.com/bketelsen/skynet/wiki/Dashboard)

##Internals
#####Query
The sky command and the client connectivity logic is all backed by [Query](skynet/wiki/Query). A struct that can be used to search the cluster for instances of services, regions, hosts, service names, service versions that Skynet is currently aware of. It's exposed for any custom need you may have for searching the cluster.

[https://github.com/bketelsen/skynet/wiki/Query](https://github.com/bketelsen/skynet/wiki/Query)

#####Instance Listener
You can create an instance listener by passing it a Query object, and be notified anytime an instance matching your Query is added/removed/changed.

[https://github.com/bketelsen/skynet/wiki/Instance-Monitor-&-Instance-Listener](https://github.com/bketelsen/skynet/wiki/Instance-Monitor-&-Instance-Listener)

## Getting Started
The [wiki](skynet/wiki) has tons of documentation and tutorials on how to get started.

The *examples/* directory has example services & clients

Also in the *examples*/ directory is a Vagrant setup with chef recipes to deploy a mock cluster using virtual machines so that you can see it in action. The wiki has a nice walkthrough on setting up and running a simulated cluster with Vagrant: [https://github.com/bketelsen/skynet/wiki/Vagrant-Example](wiki/Vagrant-Example)


## Work In Progress
##### Entire node failure / netsplit
Although a lot of work has gone into keeping instances up, and restarting, and the configuration management system to stay cleaned up. We still have a feature out standing to better handle an entire node crashing, or not being able to be communicated with.

#####Monitor CPU/Memory/Load
Skynet will be aware of system utilization and number of requests going to specific instances of a service, and will be able to have a configurable threshold for it to restart or remove itself from the pool of instances for it's particular Service/Version, and return itself to the queue when the system has leveled out, or restart has completed

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
* IRC: #skynet-dev on freenode

##Issues:
Github Issues now the canonical source of issues for Skynet.

##Open Source - MIT Software License
Copyright (c) 2012 Brian Ketelsen

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
