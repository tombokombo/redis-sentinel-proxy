package main

import (

    "net"
    "time"
    "sync"
    "strings"
    "context"
    "fmt"
    "strconv"

)

const (

     GET_SENTINELS_CMD = "sentinel sentinels"

     GET_REDISMASTER_CMD = "sentinel get-master-addr-by-name"

     NEW_WRITABLE_MASTER = "NEW_WRITABLE_MASTER"

     SAME_WRITABLE_MASTER = "SAME_WRITABLE_MASTER"

     MAX_RETRIES = 1000

)

type Sentinels struct {

    mux *sync.RWMutex

    cluster_name string

    inprogress bool

    last_sentinel_discovery time.Time

    last_redis_master_discovery time.Time

    redis_master_host string

    redis_master_writable bool

    hosts []string

    update_reties_count int

}


func (s *Sentinels) getRedisMasterAddr() string {

    s.mux.RLock()

    retval := s.redis_master_host

    s.mux.RUnlock()

    return retval

}

func (s *Sentinels) setWritable (val bool,addr string) bool {

    s.mux.RLock()

    master_addr := s.redis_master_host

    s.mux.RUnlock()

    if addr == master_addr {

        s.mux.Lock()

        s.redis_master_writable = val

        s.mux.Unlock()

        return true

    }

    return false

}

func (s *Sentinels) getWritable (addr string) bool {

    s.mux.RLock()

    master_addr := s.redis_master_host

    s.mux.RUnlock()

    if addr == master_addr {

        s.mux.RLock()

        retval := s.redis_master_writable

        s.mux.RUnlock()

        return retval

    }

    return false

}

func (s *Sentinels) setRedisMaster(redis_host string) {

        s.mux.Lock()

        s.redis_master_host = redis_host

        s.last_redis_master_discovery = time.Now()

        s.mux.Unlock()

}

func (s *Sentinels) getLastMasterDiscovery() time.Time {

    s.mux.RLock()

    retval := s.last_redis_master_discovery

    s.mux.RUnlock()

    return retval

}

func (s *Sentinels) setHosts(hosts []string) {
    
    s.mux.Lock()

    s.hosts = hosts

    s.last_sentinel_discovery = time.Now()

    s.mux.Unlock()

}

func (s *Sentinels) retry() int {

    s.mux.Lock()

    retval := s.update_reties_count

    s.update_reties_count++

    s.mux.Unlock()

    return retval

}

func (s *Sentinels) setProgress(val bool) {

    s.mux.Lock()

    s.inprogress = val

    s.mux.Unlock()

}

func (s *Sentinels) isInProgress() bool {

    s.mux.RLock()

    retval:= s.inprogress

    s.mux.RUnlock()

    return retval

}

func (s* Sentinels) getSeninelsHosts() []string {

    s.mux.RLock()

    retval := s.hosts

    s.mux.RUnlock()

    return retval
}

func (s *Sentinels) init(cnf string,cluster_name string) {

    //init could block

    _print(false,"Sentinels Init called",cnf)

    s.mux = &sync.RWMutex{}

    s.update_reties_count = 0

    s.cluster_name = cluster_name

    s.hosts = strings.Split(cnf, ",")

    s.redis_master_writable = false

    connected := false

    forloop:

    for _, host := range s.hosts {

         _print(false,"Init connecting host",host)

        conn, err := net.DialTimeout("tcp", host , time.Duration(5 * time.Second))

        if check_warn_bool(err,"Init cannot connect defined sentinel",host) {

            continue

        }

        _print(false,"Init connected to sentinel",conn.LocalAddr().String())

        s.discoverSentinels(conn)

        discovered := s.discoverRedisMaster(conn)

        conn.Close()

        if ! discovered {

            panic("init failed to get master")

        } else {

            connected = true

            break forloop

        }

        //in case that first round somehow failed give redis some time and try other conf host
        time.Sleep(1 * time.Second)

    }

    if connected {

        _print(false,"Init finished")

    } else {

        panic("not connected, init failed to get master")

    }

}


func (s *Sentinels) discoverSentinels(conn net.Conn) {

    _print(false,"discoverSentinels called")

    fmt.Fprintf(conn,  GET_SENTINELS_CMD + " " + s.cluster_name + "\n")

    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

    defer cancel()

    data := ReadConnection(ctx,conn,stopRedisResponse)

    hosts := getSentinelsFromResponse(data)

    //append actual connected host

    hosts = append(hosts, conn.RemoteAddr().String())

    _print(false,"discoverSentinels discovered", strings.Join(hosts,","))

    s.setHosts(hosts)

}

func (s *Sentinels) discoverRedisMaster(conn net.Conn) bool {

    _print(false,"discoverRedisMaster called")

    fmt.Fprintf(conn,  GET_REDISMASTER_CMD + " " + s.cluster_name + "\n")

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

    defer cancel()

    data := ReadConnection(ctx,conn,stopRedisResponse)

    redis_host, err := getRedisMasterFromResponse(data)

    _print(true,"discoverRedisMaster master found ",redis_host)

    if err != nil {

        return false

    }

    conn, err = net.DialTimeout("tcp", redis_host , time.Duration( 2 * time.Second ))

    if err != nil {

        _print(false,"discoverRedisMaster cannot connect redis master",err.Error())
        _print(false,"discoverRedisMaster should timeout if master changed, probably we are too soon")

        return false

    } else {

        defer conn.Close()

    }

    writable := isRedisMasterWritable(conn)

    if writable {

        s.setRedisMaster(redis_host)

        _ = s.setWritable(true,redis_host)

        return true

    } else {

        _ = s.setWritable(false,redis_host)

        return false

    }

}

func (s *Sentinels) UpdateRedisMaster(ch chan string,retrying bool) {

    _print(false,"UpdateRedisMaster called, retrying", strconv.FormatBool(retrying))

    redis_master_host_start := s.getRedisMasterAddr()

    if ! retrying {

        go func (redis_master_host string,chnl chan string) {

            conn, err := net.DialTimeout("tcp", redis_master_host , time.Duration(2 * time.Second))

            if err == nil {

                writable := isRedisMasterWritable(conn)

                if writable {

                    _print(false,"async test old master -> looks like old one is writable -> no update, master:",redis_master_host)

                    s.setWritable(true,redis_master_host)

                    chnl <- SAME_WRITABLE_MASTER

                    conn.Close()

                    return

                } else {

                    _print(false,"async test old master -> old master not writable, try to find new")

                    s.setWritable(false,redis_master_host)

                    conn.Close()

                }

            } else {

                _print(false,"async test old master -> old master not reachable, try to find new")

                s.setWritable(false,redis_master_host)

            }


        }(redis_master_host_start,ch)

    }

    if s.isInProgress() {

        _print(false,"UpdateRedis master in progres, ingoring")

        return

    }

    s.setProgress(true)

    hosts := s.getSeninelsHosts()

    first_win_conn_ch := make(chan net.Conn,0) 

    //don't waste time, try all at once and use first
    go getFirstConnFromArr(hosts,first_win_conn_ch)

    conn := <- first_win_conn_ch

    discovered := s.discoverRedisMaster(conn)

    conn.Close()

    if ! discovered && s.retry() < MAX_RETRIES {

        _print(false,"UpdateRedisMaster retrying, not discovered")

        time.Sleep(100* time.Millisecond)

        s.setProgress(false)

        go s.UpdateRedisMaster(ch,true)

        return

    } else if ! discovered {

        panic("cannot discover redis master and max retries reached")

    }

    redis_master_host := s.getRedisMasterAddr()

    s.setProgress(false)

    if redis_master_host_start != redis_master_host {

        ch <- NEW_WRITABLE_MASTER

    } else {

        ch <- SAME_WRITABLE_MASTER

    }

}
