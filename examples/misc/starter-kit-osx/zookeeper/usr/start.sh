#!/bin/sh
base_dir=$(dirname $0)
for file in $base_dir/libs/*.jar;
do
  CLASSPATH=$CLASSPATH:$file
done
CLASSPATH=$CLASSPATH:config
java -cp $CLASSPATH \
	org.apache.zookeeper.server.quorum.QuorumPeerMain \
	$base_dir/config/zoo-air.cfg
