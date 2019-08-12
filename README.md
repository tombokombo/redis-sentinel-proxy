# Redis-Sentinel-Proxy

Highly inspired by https://github.com/metal3d/redis-ellison/ , but smarter :)

Proxy for apps using redis which for some reasons cannot interact with redis sentinels cluster.

# Usage

## Start Cluster with test client and tests
You will need two terminals, one running proxy and test client

```
./start_cluster.sh
```
Second one running tests which will kill/start redis and sentinel instancies

```
./start_tests.sh
```

## Build binary from source


#TODO
test redis pubsub functionality as well
