# DESCRIPTION:

Installs and configures MongoDB, supporting:

* Single MongoDB
* Replication
* Sharding
* Replication and Sharding
* 10gen repository package installation

# REQUIREMENTS:

## Platform:

The cookbook aims to be platform independant, but is best tested on debian squeeze systems.

The `10gen_repo` recipe currently supports only the Debian and Ubuntu 10gen repository.
Patches for other platforms are welcome.

# DEFINITIONS:

This cookbook contains a definition `mongodb_instance` which can be used to configure
a certain type of mongodb instance, like the default mongodb or various components
of a sharded setup.

For examples see the USAGE section below.

# ATTRIBUTES:

* `mongodb[:dbpath]` - Location for mongodb data directory, defaults to "/var/lib/mongodb"
* `mongodb[:logpath]` - Path for the logfiles, default is "/var/log/mongodb"
* `mongodb[:port]` - Port the mongod listens on, default is 27017
* `mongodb[:client_role]` - Role identifing all external clients which should have access to a mongod instance
* `mongodb[:cluster_name]` - Name of the cluster, all members of the cluster must
    reference to the same name, as this name is used internally to identify all
    members of a cluster.
* `mongodb[:shard_name]` - Name of a shard, default is "default"
* `mongodb[:sharded_collections]` - Define which collections are sharded
* `mongodb[:replicaset_name]` - Define name of replicatset

# USAGE:

## 10gen repository

Adds the stable [10gen repo](http://www.mongodb.org/downloads#packages) for the
corresponding platform. Currently only implemented for the Debian and Ubuntu repository.

Usage: just add `recipe[mongodb::10gen_repo]` to the node run_list *before* any other
MongoDB recipe, and the mongodb-10gen **stable** packages will be installed instead of the distribution default.

## Single mongodb instance

Simply add

```ruby
include_recipe "mongodb::default"
```
  
to your recipe. This will run the mongodb instance as configured by your distribution.
You can change the dbpath, logpath and port settings (see ATTRIBUTES) for this node by
using the `mongodb_instance` definition:

```ruby
mongodb_instance "mongodb" do
  port node['application']['port']
end
```

This definition also allows you to run another mongod instance with a different
name on the same node

```ruby
mongodb_instance "my_instance" do
  port node['mongodb']['port'] + 100
  dbpath "/data/"
end
```
  
The result is a new system service with

```shell
  /etc/init.d/my_instance <start|stop|restart|status>
```

## Replicasets

Add `mongodb::replicaset` to the node's run_list. Also choose a name for your
replicaset cluster and set the value of `node[:mongodb][:cluster_name]` for each
member to this name.

## Sharding

You need a few more components, but the idea is the same: identification of the
members with their different internal roles (mongos, configserver, etc.) is done via
the `node[:mongodb][:cluster_name]` and `node[:mongodb][:shard_name]` attributes.

Let's have a look at a simple sharding setup, consisting of two shard servers, one
config server and one mongos.

First we would like to configure the two shards. For doing so, just use
`mongodb::shard` in the node's run_list and define a unique `mongodb[:shard_name]`
for each of these two nodes, say "shard1" and "shard2".

Then configure a node to act as a config server - by using the `mongodb::configserver`
recipe.

And finally you need to configure the mongos. This can be done by using the
`mongodb::mongos` recipe. The mongos needs some special configuration, as these
mongos are actually doing the configuration of the whole sharded cluster.
Most importantly you need to define what collections should be sharded by setting the
attribute `mongodb[:sharded_collections]`:

```javascript
{
  "mongodb": {
    "sharded_collections": {
      "test.addressbook": "name",
      "mydatabase.calendar": "date"
    }
  }
}
```
  
Now mongos will automatically enable sharding for the "test" and the "mydatabase"
database. Also the "addressbook" and the "calendar" collection will be sharded,
with sharding key "name" resp. "date".
In the context of a sharding cluster always keep in mind to use a single role
which is added to all members of the cluster to identify all member nodes.
Also shard names are important to distinguish the different shards.
This is esp. important when you want to replicate shards.

## Sharding + Replication

The setup is not much different to the one described above. All you have to do is adding the 
`mongodb::replicaset` recipe to all shard nodes, and make sure that all shard
nodes which should be in the same replicaset have the same shard name.

For more details, you can find a [tutorial for Sharding + Replication](https://github.com/edelight/chef-cookbooks/wiki/MongoDB%3A-Replication%2BSharding) in the wiki.

# LICENSE and AUTHOR:

Author:: Markus Korn <markus.korn@edelight.de>

Copyright:: 2011, edelight GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
