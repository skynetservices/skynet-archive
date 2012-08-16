#
# Cookbook Name:: mongodb
# Definition:: mongodb
#
# Copyright 2011, edelight GmbH
# Authors:
#       Markus Korn <markus.korn@edelight.de>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

define :mongodb_instance, :mongodb_type => "mongod" , :action => [:enable, :start], :port => 27017 , \
    :logpath => "/var/log/mongodb", :dbpath => "/data", :configfile => "/etc/mongodb.conf", \
    :configserver => [], :replicaset => nil, :enable_rest => false, \
    :notifies => [] do
    
  include_recipe "mongodb::default"
  
  name = params[:name]
  type = params[:mongodb_type]
  service_action = params[:action]
  service_notifies = params[:notifies]
  
  port = params[:port]

  logpath = params[:logpath]
  logfile = "#{logpath}/#{name}.log"
  
  dbpath = params[:dbpath]
  
  configfile = params[:configfile]
  configserver_nodes = params[:configserver]
  
  replicaset = params[:replicaset]
  if type == "shard"
    if replicaset.nil?
      replicaset_name = nil
    else
      # for replicated shards we autogenerate the replicaset name for each shard
      replicaset_name = "rs_#{replicaset['mongodb']['shard_name']}"
    end
  else
    # if there is a predefined replicaset name we use it,
    # otherwise we try to generate one using 'rs_$SHARD_NAME'
    begin
      replicaset_name = replicaset['mongodb']['replicaset_name']
    rescue
      replicaset_name = nil
    end
    if replicaset_name.nil?
      begin
        replicaset_name = "rs_#{replicaset['mongodb']['shard_name']}"
      rescue
        replicaset_name = nil
      end
    end
  end
  
  if !["mongod", "shard", "configserver", "mongos"].include?(type)
    raise "Unknown mongodb type '#{type}'"
  end
  
  if type != "mongos"
    daemon = "/usr/bin/mongod"
    configserver = nil
    configfile = nil
    Chef::Log.warn("We are not using a configfile, as the daemons can be configured via commandline")
  else
    daemon = "/usr/bin/mongos"
    configfile = nil
    dbpath = nil
    configserver = configserver_nodes.collect{|n| "#{n['fqdn']}:#{n['mongodb']['port']}" }.join(",")
  end
  
  # default file
  template "#{node['mongodb']['defaults_dir']}/#{name}" do
    action :create
    source "mongodb.default.erb"
    group node['mongodb']['root_group']
    owner "root"
    mode "0644"
    variables(
      "daemon_path" => daemon,
      "name" => name,
      "config" => configfile,
      "configdb" => configserver,
      "port" => port,
      "logpath" => logfile,
      "dbpath" => dbpath,
      "replicaset_name" => replicaset_name,
      "configsrv" => false, #type == "configserver", this might change the port
      "shardsrv" => false,  #type == "shard", dito.
      "enable_rest" => params[:enable_rest]
    )
    notifies :restart, "service[#{name}]"
  end
  
  # log dir [make sure it exists]
  directory logpath do
    owner "mongodb"
    group "mongodb"
    mode "0755"
    action :create
    recursive true
  end
  
  if type != "mongos"
    # dbpath dir [make sure it exists]
    directory dbpath do
      owner "mongodb"
      group "mongodb"
      mode "0755"
      action :create
      recursive true
    end
  end
  
  # init script
  template "#{node['mongodb']['init_dir']}/#{name}" do
    action :create
    source "mongodb.init.erb"
    group node['mongodb']['root_group']
    owner "root"
    mode "0755"
    variables :provides => name
    notifies :restart, "service[#{name}]"
  end
  
  # service
  service name do
    supports :status => true, :restart => true
    action service_action
    notifies service_notifies
    if !replicaset_name.nil?
      notifies :create, "ruby_block[config_replicaset]"
    end
    if type == "mongos"
      notifies :create, "ruby_block[config_sharding]", :immediately
    end
    if name == "mongodb"
      # we don't care about a running mongodb service in these cases, all we need is stopping it
      ignore_failure true
    end
  end
  
  # replicaset
  if !replicaset_name.nil?
    rs_nodes = search(
      :node,
      "mongodb_cluster_name:#{replicaset['mongodb']['cluster_name']} AND \
       recipes:mongodb\\:\\:replicaset AND \
       mongodb_shard_name:#{replicaset['mongodb']['shard_name']} AND \
       chef_environment:#{replicaset.chef_environment}"
    )
  
    ruby_block "config_replicaset" do
      block do
        if not replicaset.nil?
          MongoDB.configure_replicaset(replicaset, replicaset_name, rs_nodes)
        end
      end
      action :nothing
    end
  end
  
  # sharding
  if type == "mongos"
    # add all shards
    # configure the sharded collections
    
    shard_nodes = search(
      :node,
      "mongodb_cluster_name:#{node['mongodb']['cluster_name']} AND \
       recipes:mongodb\\:\\:shard AND \
       chef_environment:#{node.chef_environment}"
    )
    
    ruby_block "config_sharding" do
      block do
        if type == "mongos"
          MongoDB.configure_shards(node, shard_nodes)
          MongoDB.configure_sharded_collections(node, node['mongodb']['sharded_collections'])
        end
      end
      action :nothing
    end
  end
end

