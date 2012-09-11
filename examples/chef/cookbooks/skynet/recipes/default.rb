#
# Cookbook Name:: skynet
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
    go get github.com/kballard/go-shellquote && go get github.com/sbinet/liner && go get github.com/bketelsen/skynet
  }

  not_if do
    File.exists?("/opt/local/gopath/src/github.com/bketelsen/skynet")
  end
end

execute "set-environment-variables" do
  ENV['SKYNET_REGION'] = node[:skynet_region]
  ENV['SKYNET_DZHOST'] = node[:skynet_dzhost]
  ENV['SKYNET_DZNSHOST'] = node[:skynet_dznshost]
  ENV['SKYNET_DZDISCOVER'] = node[:skynet_dzdizcover]
  ENV['SKYNET_BIND_IP'] = node[:skynet_bind_ip]
  ENV['SKYNET_MIN_PORT'] = node[:skynet_min_port]
  ENV['SKYNET_MAX_PORT'] = node[:skynet_max_port]
  ENV['SKYNET_MGOSERVER'] = node[:skynet_mgoserver]
  ENV['SKYNET_MGODB'] = node[:skynet_mgoserver]


  command %Q{
    echo "export SKYNET_REGION=#{node[:skynet_region]}\nexport SKYNET_DZHOST=#{node[:skynet_dzhost]}\nexport SKYNET_DZNSHOST=#{node[:skynet_dznshost]}\nexport SKYNET_DZDISCOVER=#{node[:skynet_dzdizcover]}\nexport SKYNET_BIND_IP=#{node[:skynet_bind_ip]}\nexport SKYNET_MIN_PORT=#{node[:skynet_min_port]}\nexport SKYNET_MAX_PORT=#{node[:skynet_max_port]}\nexport SKYNET_MGOSERVER=#{node[:skynet_mgoserver]}\nexport SKYNET_MGODB=#{node[:skynet_mgodb]}" > /etc/profile.d/skynet_env.sh
  }

  not_if "ls /etc/profile.d/skynet_env.sh"
end

execute "update-skynet" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet'

  branch = node[:skynet_branch] || "master"

  command %Q{
    git fetch && git checkout #{branch} && git pull origin #{branch}
  }
end
execute "rebuild-sky" do
  cwd '/opt/local/gopath/bin'

  command %Q{
    rm sky
  }

  not_if do
    node[:skynet_rebuild] != true || !File.exists?("/opt/local/gopath/bin") || !File.exists?("/opt/local/gopath/bin/sky")
  end
end

execute "install-sky" do
  cwd '/opt/local/gopath/src/github.com/bketelsen/skynet/cmd/sky'

  command %Q{
    go install  
  }

  not_if do
    File.exists?("/opt/local/gopath/bin/sky")
  end
end
