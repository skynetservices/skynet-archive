#
# Cookbook Name:: skynet
# Recipe:: fibservice
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

execute "rebuild-fibservice" do
  cwd '/opt/local/gopath/bin'

  command %Q{
    rm fibservice
  }

  not_if do
    node[:skynet_rebuild] != true || !File.exists?("/opt/local/gopath/bin") || !File.exists?("/opt/local/gopath/bin/fibservice")
  end
end

execute "install-fibonacci-service" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet/examples/testing/fibonacci/fibservice'

  command %Q{
    go install  
  }

  not_if do
   File.exist?("/opt/local/gopath/bin/fibservice")
  end
end
