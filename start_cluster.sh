#!/bin/bash

TESTDIR=tests
CONF_DIR=${CONF_DIR:-./configs}
TEMPL_DIR=${CONF_DIR}/templates

TAIL=true
while (( "$#" )); do
  case "$1" in
    "notail")
      TAIL=false
      shift
      ;;
    *) # preserve positional arguments
      PARAMS="$PARAMS $1"
      shift
      ;;
  esac
done

eval set -- "$PARAMS"
export PREFIX=$1

trap ctrl_c INT

( ! [ $EUID = 0 ] ) && SUDO="sudo -E" || SUDO=

function ctrl_c() {
    echo down AND exit
    sudo docker-compose down && exit 0
    rm ./configs/*conf
}

function prepare_configs() {

    for i in $(seq 1 5);do
        cp $CONF_DIR/templates/sentinel.tmpl $CONF_DIR/sent_$i.conf 
        cp $CONF_DIR/templates/redis.tmpl $CONF_DIR/red_$i.conf
        if [[ $i -gt 1 ]];then
            echo slaveof redis1 6379 >> $CONF_DIR/red_$i.conf
        fi
    done
    $SUDO chmod 777 $CONF_DIR/*.conf

}

cd $TESTDIR
prepare_configs

$SUDO docker-compose up -d --build

if $TAIL ;then 
    $SUDO docker-compose logs -f redis-sentinel-proxy client
fi
