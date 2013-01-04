#
# Cookbook Name:: go
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

execute "download-doozerd" do
  command %Q{
    go get github.com/skynetservices/goprotobuf/proto && go get github.com/skynetservices/pretty && go get github.com/skynetservices/doozer && go get github.com/skynetservices/doozerd
  }

  not_if "ls $GOPATH/src/github.com/skynetservices/doozerd"
end

execute "install-doozer" do
  cwd '/opt/local/gopath/src/github.com/skynetservices/doozer/cmd/doozer'

  command %Q{
    go install  
  }
end

execute "install-doozerd" do
  cwd '/opt/local/gopath/src/github.com/skynetservices/doozerd'

  command %Q{
    go install  
  }
end
