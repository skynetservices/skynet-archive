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

execute "install-go" do
  command %Q{
    cd /opt/local && mkdir gopath && hg clone -u release https://code.google.com/p/go && cd go/src && ./make.bash
  }
  not_if "ls /opt/local/go/bin/go"
end

execute "set-go-paths" do

  goroot = '/opt/local/go'
  gopath = '/opt/local/gopath'
  path = '/opt/local/go/bin:/opt/local/gopath/bin'

  Dir[File.join('/', 'opt', 'hostgopaths', '*')].count { |file|
    gopath += ":#{file}"
    path += ":#{file}/bin"
  }

  ENV['GOPATH'] = gopath
  ENV['GOROOT'] = goroot
  ENV['PATH'] = "#{ENV['PATH']}:#{path}"


  command %Q{
    echo "export GOPATH=#{gopath}\nexport GOROOT=#{goroot}\nexport PATH=$PATH:#{path}" > /etc/profile.d/go_env.sh
  }

  not_if "ls /etc/profile.d/go_env.sh"
end
