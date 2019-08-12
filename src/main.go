package main

import (
    "github.com/namsral/flag"
    "net"
    "context"
    "time"
)

const (

    MASTER_CHANGED = "MASTER_CHANGED"

)

var (

    // defaults

    cfg_listen = "127.0.0.1:61379"

    cfg_sentinels = "127.0.0.1:26379"

    cfg_verbose = true

    cfg_debug = true

    cfg_cluster_name = "mymaster"

    cfg_cluster_convergence_maxwaittime = 30

)

func main() {

    flag.StringVar(&cfg_listen, "listen", cfg_listen,
      "listen on addr:port")

    flag.BoolVar(&cfg_verbose, "verbose", cfg_verbose,
        "be verbose")

    flag.BoolVar(&cfg_debug, "debug", cfg_debug,
        "be even more verbose")

    flag.StringVar(&cfg_sentinels, "sentinels", cfg_sentinels,
        "sentinels, host:ip separated by comma ")

    flag.StringVar(&cfg_cluster_name, "clustername", cfg_cluster_name,
        "redis cluster name")

    flag.IntVar(&cfg_cluster_convergence_maxwaittime, "maxtwaitime", cfg_cluster_convergence_maxwaittime,
        "senconds to wait since cluster state change to converge")

    flag.Parse()

    listener, err := net.Listen("tcp", cfg_listen)

    check_strict(err)

    _print(true,"listening on ", cfg_listen)

    snts := new(Sentinels)

    snts.init(cfg_sentinels,cfg_cluster_name)

    notify_chan := make(chan string,1000)

    clients := new(Clients)

    clients.init()

    go clients.receiver(notify_chan,snts)

    for {

            conn, err := listener.Accept()

            if check_warn_bool(err) { continue }

            go func(cl *Clients,sl *Sentinels, notify_chan chan string) {

                rec := make(chan string)

                go sl.UpdateRedisMaster(rec,false)

                ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg_cluster_convergence_maxwaittime) * time.Second)

                defer cancel()

                select {

                    case msg := <- rec:
                        //forward msg to global
                        notify_chan <- msg

                        master := sl.getRedisMasterAddr()

                        _print(true,"proxying new client, to master addr ", master)
                        go cl.proxy(conn,master)

                    case <-ctx.Done():
                        //timeout reject client
                        _print(false,"master info too late, closing client")

                        conn.Close()

                }

            }(clients,snts,notify_chan)

            _print(true,"Accepted",conn.RemoteAddr().String())

    }

}
