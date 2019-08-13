# Redis-Sentinel-Proxy

Highly inspired by https://github.com/metal3d/redis-ellison/ , but smarter :) You should read his description which fits this project. This incarnation is event based and testing redis to know if redis master is writable/ready to use.

Proxy for apps using redis which for some reasons cannot interact with redis sentinels cluster.


# Usage

## Start Cluster with test client and tests
You will need two terminals, one running proxy and test client

```
./start_cluster.sh
```
Second terminal for running tests which will kill/start redis and sentinel instancies, there are different test included, without arguments, script will run just basic test

```
./start_tests.sh
```

## example for all tests with containers names prefix

```
./start_cluster.sh myprefix_
```

```
./start_tests.sh all myprefix_
```

## Build binary from source


#TODO
test redis pubsub functionality as well
