#
# Cookbook Name:: skynet-example
# Recipe:: default
#
# Copyright 2012, Erik St. Martin
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

directory "/opt/local" do
  mode 0755
  recursive true
end

execute "download-skynet" do
  command %Q{
    go get github.com/bketelsen/skynet
  }

  not_if "ls $GOPATH/src/github.com/bketelsen/skynet"
end

execute "update-skynet" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet'

  branch = node[:skynet_branch] || "master"

  command %Q{
    git checkout #{branch} && git pull origin #{branch}
  }
end

execute "forced-rebuild" do
 cwd '/opt/local/gopath/bin'

 if node[:skynet_rebuild] == true
   command %Q{
     rm fibservice service
   }
 end
end

execute "install-example-service" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet/examples/service'

  command %Q{
    go install  
  }

  not_if "ls $GOPATH/bin/service"
end

execute "install-fibonacci-service" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet/examples/testing/fibonacci/fibservice'

  command %Q{
    go install  
  }

  not_if "ls $GOPATH/bin/fibservice"
end
