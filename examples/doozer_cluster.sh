# Start a local cluster of doozers for testings
#
# Usage: doozer_cluster.sh <start|stop> (<instances> <dzns_instances>)

BIND_IP="127.0.0.1"

START_PORT=8046
START_WEB_PORT=8080

START_DZNS_PORT=10000
START_DZNS_WEB_PORT=11000

DZNS_INSTANCES=5
CLUSTER_INSTANCES=5

if [ $# -lt 1 ] 
then
  echo "Usage: doozer_cluster.sh <start|stop> (<instances> <dzns_instances>)"
  exit
fi

echo "Using Doozerd: $DOOZERD_PATH"

function start {
  if [ $# -gt 1 ]; then
    CLUSTER_INSTANCES=$3
  fi

  if [ $# -eq 3 ]; then
    DZNS_INSTANCES=$3
  fi

  # First startup our DzNS cluster
  dzns_count=0
  dzns_port=$START_DZNS_PORT
  dzns_web_port=$START_DZNS_WEB_PORT

  until [ $dzns_count -eq $DZNS_INSTANCES ]
  do

    if [ $dzns_port != $START_DZNS_PORT ]
    then
      echo doozerd -timeout 5 -l "$BIND_IP:$dzns_port" -w ":$dzns_web_port" -c "dzns" -a "$BIND_IP:$START_DZNS_PORT"
      doozerd -timeout 5 -l "$BIND_IP:$dzns_port" -w ":$dzns_web_port" -c "dzns" -a "$BIND_IP:$START_DZNS_PORT" 2>/dev/null &

      # add to DzNS cluster
     echo "\c" | doozer -a "doozer:?ca=$BIND_IP:$START_DZNS_PORT" add "/ctl/cal/$dzns_count" >/dev/null &
    else
      echo doozerd -l "$BIND_IP:$dzns_port" -w ":$dzns_web_port" -c "dzns"
      doozerd -l "$BIND_IP:$dzns_port" -w ":$dzns_web_port" -c "dzns" 2>/dev/null &
      sleep 1
    fi

    dzns_port=$(( $dzns_port + 1 ))
    dzns_web_port=$(( $dzns_web_port + 1 ))
    dzns_count=$(( $dzns_count + 1 ))
  done

  # Now startup doozer instances
  dz_count=0
  dz_port=$START_PORT
  dz_web_port=$START_WEB_PORT

  until [ $dz_count -eq $CLUSTER_INSTANCES ]
  do
    echo doozerd -timeout 5 -l "$BIND_IP:$dz_port" -w ":$dz_web_port" -c "skynet" -b "doozer:?ca=$BIND_IP:$START_DZNS_PORT"
    doozerd -timeout 5 -l "$BIND_IP:$dz_port" -w ":$dz_web_port" -c "skynet" -b "doozer:?ca=$BIND_IP:$START_DZNS_PORT" 2>/dev/null &

    if [ $dz_port != $START_PORT ]
    then
      # add to cluster
      # this has to connect to master, it blows up with a REV_MISMATCH if we connect to anyone else
      echo -n | doozer -a "doozer:?ca=$BIND_IP:$START_PORT" -b "doozer:?ca$BIND_IP:$START_DZNS_PORT" add "/ctl/cal/$dz_count" >/dev/null &
    else
      sleep 1
    fi

    dz_port=$(( $dz_port + 1 ))
    dz_web_port=$(( $dz_web_port + 1 ))
    dz_count=$(( $dz_count + 1 ))
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
