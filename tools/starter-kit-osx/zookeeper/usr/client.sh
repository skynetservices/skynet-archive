#!/bin/sh
base_dir=$(dirname $0)
for file in $base_dir/libs/*.jar;
do
  CLASSPATH=$CLASSPATH:$file
done
CLASSPATH=$CLASSPATH:config
java -cp $CLASSPATH \
	org.apache.zookeeper.ZooKeeperMain -server 127.0.0.1:2181
