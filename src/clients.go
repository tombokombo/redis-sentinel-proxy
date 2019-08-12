package main

import (

    "net"
    "sync"
    "time"

)

type Clients struct {

    mux *sync.Mutex

    connections []*Connection
}

type Connection struct {

    connected time.Time

    conn net.Conn

    redis_master string

    closed bool
}

func (connect *Connection) Close() {

    _print(false,"client Closed")

    connect.conn.Close()

    connect.closed = true

}

func (c *Clients) init() {

    c.mux = &sync.Mutex{}

}

func (c *Clients) addClient(conn net.Conn, host string) *Connection {

    cn := new(Connection)

    cn.connected = time.Now()

    cn.conn = conn

    cn.redis_master = host

    cn.closed = false

    c.mux.Lock()

    c.connections = append(c.connections, cn)

    c.mux.Unlock()

    return cn

}

func (c *Clients) receiver(ch chan string, sl *Sentinels) {

    for msg := range ch {

        if msg == NEW_WRITABLE_MASTER {

            last_discovery := sl.getLastMasterDiscovery()

            redis_master := sl.getRedisMasterAddr()

            connected := make([]*Connection,0)

            for idx, cn := range c.connections {

                if  ( cn.connected.Before(last_discovery) && redis_master != cn.redis_master ) || cn.closed {

                  if sl.getWritable(redis_master) {
                  
                    _print(true,"flushing old client", "found master",
                        redis_master , "client used master", cn.redis_master,"client conn:=>", cn.conn.RemoteAddr().String(),cn.conn.LocalAddr().String())

                    cn.conn.Close()

                  } else {

                      _print(false,"ALMOST flush old client but muster not writable",
                          redis_master , "client used master", cn.redis_master,"client conn:=>", cn.conn.RemoteAddr().String(),cn.conn.LocalAddr().String())

                      connected = append(connected, c.connections[idx])

                  }

                } else {

                    connected = append(connected, c.connections[idx])

                }

            }

            c.mux.Lock()

            c.connections =  connected

            c.mux.Unlock()

        }

    }

}

func (c *Clients) proxy(client_conn net.Conn,master_addr string) {

    redis_conn, err := net.DialTimeout("tcp", master_addr, time.Duration(2 * time.Second))

    if err == nil {

        cn := c.addClient(client_conn,master_addr)

        go dataCopyToClient(redis_conn, cn)

        go dataCopyFromClient(cn, redis_conn)

    } else {

        client_conn.Close()

    }

}

