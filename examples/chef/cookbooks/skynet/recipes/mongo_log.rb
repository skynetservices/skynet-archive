#
# Cookbook Name:: skynet
# Recipe:: mongo_log
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

GIGABYTE = 1073741824
COLLECTION_SIZE = GIGABYTE * 1

execute "create-logging-database" do
  command %Q{
    mongo skynet --eval "db.createCollection('log', {capped:true, size:#{COLLECTION_SIZE}})"
  }

  not_if do
    output = `mongo skynet --eval "db.getCollectionNames()"`
    collections = output.split("\n").last.split(",")

    collections.include?("log")
  end
end
