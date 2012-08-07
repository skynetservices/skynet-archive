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
    go get code.google.com/p/gonicetrace/nicetrace && go get github.com/bketelsen/skynet
  }

  #not_if "ls $GOPATH/src/github.com/bketelsen/skynet"
end

execute "install-example-service" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet/examples/service'

  command %Q{
    go install  
  }
end
