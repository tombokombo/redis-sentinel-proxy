#!/bin/bash
TESTDIR=tests
EXAMPLE="e.g ./start_tests.sh basic myprefix_"

[[ $# -eq 0 ]] && echo -e \
"no params, args are test names [basic] [partition] [random] [all] and optional compose prefix\n\
$EXAMPLE
starting basic test" && BASIC=True

while (( "$#" )); do
  case "$1" in
   "basic")
      echo "basic test"
      BASIC=True
      shift
      ;;
   "partition")
      echo "net partioning test"
      NET_PARTITIONING=True
      shift
      ;;
   "random")
      echo "random redis/sentinel down"
      RAND=True
      shift
      ;;
   "all")
      echo "performing all tests"
      ALL=True
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

( ! [ $EUID = 0 ] ) && SUDO="sudo -E" || SUDO=


function basic() {
    for i in `seq 1 5`;do 
        $SUDO docker-compose kill redis${i};
        sleep 2;
        $SUDO docker-compose start redis${i} || exit 1;
        sleep 5;done
}

function random() {
    for i in `seq 1 200`;do
        rnd=$(($RANDOM%5+1))
        echo docker-compose kill redis${rnd}
        $SUDO docker-compose kill redis${rnd}
        sleep 5
        echo docker-compose start redis${rnd}
        $SUDO docker-compose start redis${rnd} || exit 1;
    done
}

function net_partitioning() {

    for a in `seq 1 10`;do
        s=1 e=2 
        (( $a % 2 )) && s=3 e=5
        for i in `seq 1 6`; do
            for j in `seq $s $e`;do 
                action=start
                (( $i % 2 )) && action=kill
                echo docker-compose $action redis${j}
                $SUDO docker-compose $action redis${j} || exit 1
                echo docker-compose $action sentinel${j}
                $SUDO docker-compose $action sentinel${j} || exit 1
            done
            sleep 10
        done
    done
}

cd $TESTDIR

if [ $BASIC ] ;then
    basic
fi

if [ $NET_PARTITIONING ]; then
    net_partitioning
fi

if [ $RAND ]; then
    random
fi

if [ $ALL ]; then
    basic
    net_partitioning
    random
fi

[ ! $BASIC ] &&  [ ! $NET_PARTITIONING ] &&  [ ! $RAND ] && [ ! $ALL ] && echo "missing test arg $EXAMPLE"
