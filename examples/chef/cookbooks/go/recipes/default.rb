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
    cd /opt/local && mkdir gopath && hg clone -u release https://code.google.com/p/go && cd go/src && ./all.bash
  }
  not_if "ls /opt/local/go/bin/go"
end

execute "set-go-paths" do
  ENV['GOPATH'] = '/opt/local/gopath'
  ENV['GOROOT'] = '/opt/local/go'
  ENV['PATH'] = "#{ENV['PATH']}:/opt/local/go/bin"

  goroot = '/opt/local/go'
  gopath = '/opt/local/gopath'
  path = '/opt/local/go/bin:/opt/local/gopath/bin'

  gps = ENV['GOPATH'].split(':')
  gpi = 0
  gps.each do |gp|
    gopath += ":/opt/hostgopaths/gp#{gpi}"
    path += ":/opt/hostgopaths/gp#{gpi}/bin"
    gpi += 1
  end


  command %Q{
    echo "GOPATH=#{gopath}\nGOROOT=#{goroot}\nPATH=$PATH:#{path}" > /etc/profile.d/go_env.sh
  }

  #not_if "ls /etc/profile.d/go_env.sh"
end
