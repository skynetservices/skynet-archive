# Start a local cluster of doozers for testings
#
# Usage: doozer_cluster.sh <start|stop|restart> [<instances> <dzns_instances>]

BIND_IP="127.0.0.1"

START_PORT=8046
START_WEB_PORT=8080

START_DZNS_PORT=10000
START_DZNS_WEB_PORT=11000

DZNS_INSTANCES=5
CLUSTER_INSTANCES=5

function usage {
  echo "Usage: doozer_cluster.sh <start|stop|restart> [<instances> <dzns_instances>]"
  exit 1
}

if [ $# -lt 1 ] 
then
  usage
fi

if [ $1 != "start" ] && [ $1 != "stop" ] && [ $1 != "restart" ]
then
  usage
fi

if [ $# -gt 1 ]; then
  CLUSTER_INSTANCES=$2
  if [ $CLUSTER_INSTANCES -lt 1 ]
  then
    echo "Need at least one instance (have $CLUSTER_INSTANCES)"
    exit 1
  fi
fi

if [ $# -eq 3 ]; then
  DZNS_INSTANCES=$3
  if [ $DZNS_INSTANCES -lt 1 ]
  then
    echo "Need at least one DZNS instance (have $DZNS_INSTANCES)"
    exit 1
  fi
fi


echo "Using Doozerd: $DOOZERD_PATH"

function start {
  # First startup our DzNS cluster
  for dzns_count in $(seq 0 $(($DZNS_INSTANCES-1)))
  do
    dzns_port=$((START_DZNS_PORT+$dzns_count))
    dzns_web_port=$((START_DZNS_WEB_PORT+$dzns_count))
    echo doozerd -timeout 5 -l "$BIND_IP:$dzns_port" -w ":$dzns_web_port" -c "dzns" -a "$BIND_IP:$START_DZNS_PORT"
    doozerd -timeout 5 -l "$BIND_IP:$dzns_port" -w ":$dzns_web_port" -c "dzns" -a "$BIND_IP:$START_DZNS_PORT" 2>/dev/null &

    if [ $dzns_count -eq 0 ]
    then
      sleep 1
    else
      # add to DzNS cluster
      echo "\c" | doozer -a "doozer:?ca=$BIND_IP:$START_DZNS_PORT" add "/ctl/cal/$dzns_count" >/dev/null &
    fi
  done

  # Now startup doozer instances
  for dz_count in $(seq 0 $(($CLUSTER_INSTANCES-1)))
  do
    dz_port=$(( $START_PORT + $dz_count ))
    dz_web_port=$(( $START_WEB_PORT + $dz_count ))
    echo doozerd -timeout 5 -l "$BIND_IP:$dz_port" -w ":$dz_web_port" -c "skynet" -b "doozer:?ca=$BIND_IP:$START_DZNS_PORT"
    doozerd -timeout 5 -l "$BIND_IP:$dz_port" -w ":$dz_web_port" -c "skynet" -b "doozer:?ca=$BIND_IP:$START_DZNS_PORT" 2>/dev/null &

    if [ $dzns_count -eq 0 ]
    then
      sleep 1
    else
      # add to cluster
      # this has to connect to master, it blows up with a REV_MISMATCH if we connect to anyone else
      echo -n | doozer -a "doozer:?ca=$BIND_IP:$START_PORT" -b "doozer:?ca$BIND_IP:$START_DZNS_PORT" add "/ctl/cal/$dz_count" >/dev/null &
    fi
  done
}

function stop {
  killall doozerd
}

if [ $1 = "start" ]; then
  start
elif [ $1 = "stop" ]; then
  stop
elif [ $1 = "restart" ]; then
  stop
  start
fi
