#!/bin/sh
ps ax | grep -i 'org.apache.zookeeper.server.quorum.QuorumPeerMain' | grep -v grep | awk '{print $1}' | xargs kill -SIGINT
