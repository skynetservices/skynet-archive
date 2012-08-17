maintainer        "edelight GmbH"
maintainer_email  "markus.korn@edelight.de"
license           "Apache 2.0"
description       "Installs and configures mongodb"
version           "0.11"

recipe "mongodb", "Installs and configures a single node mongodb instance"
recipe "mongodb::10gen_repo", "Adds the 10gen repo to get the latest packages"
recipe "mongodb::mongos", "Installs and configures a mongos which can be used in a sharded setup"
recipe "mongodb::configserver", "Installs and configures a configserver for mongodb sharding"
recipe "mongodb::shard", "Installs and configures a single shard"
recipe "mongodb::replicaset", "Installs and configures a mongodb replicaset"

depends "apt"

%w{ ubuntu debian freebsd }.each do |os|
  supports os
end

attribute "mongodb/dbpath",
  :display_name => "dbpath",
  :description => "Path to store the mongodb data",
  :default => "/var/lib/mongodb"
  
attribute "mongodb/logpath",
  :display_name => "logpath",
  :description => "Path to store the logfiles of a mongodb instance",
  :default => "/var/log/mongodb"
  
attribute "mongodb/port",
  :display_name => "Port",
  :description => "Port the mongodb instance is running on",
  :default => "27017"
  
attribute "mongodb/client_roles",
  :display_name => "Client Roles",
  :description => "Roles of nodes who need access to the mongodb instance",
  :default => []
  
attribute "mongodb/cluster_name",
  :display_name => "Cluster Name",
  :description => "Name of the mongodb cluster, all nodes of a cluster must have the same name.",
  :default => nil

attribute "mongodb/shard_name",
  :display_name => "Shard name",
  :description => "Name of a mongodb shard",
  :default => "default"  
  
attribute "mongodb/sharded_collections",
  :display_name => "Sharded Collections",
  :description => "collections to shard",
  :default => {}

attribute "mongodb/replicaset_name",
  :display_name => "Replicaset_name",
  :description => "Name of a mongodb replicaset",
  :default => nil
  
attribute "mongodb/enable_rest",
  :display_name => "Enable Rest",
  :description => "Enable the ReST interface of the webserver"
