package main

import (

    "net"
    "strings"
    "strconv"
    "fmt"
    "context"
    "time"
    "errors"

)

const (

    REDIS_READONLY = "-READONLY"

    REDIS_DEL_KEY = "unikatnykurvaklucktoryurciteniktonema"

    DEL_REDIS_CMD = "DEL"

)



func stopRedisResponse(stop chan bool,data chan string) {

    _print(false,"stopRedisResponse called")
    //for nested array
    array_size := 0

    line_before := ""

    for {

      select {

          case line := <-data:

              if strings.HasPrefix(line,"-ERR") || ( strings.Compare(line,":0") == 0 ) {

                  stop <- true

                  close(stop)

                  return

              }

              if strings.HasPrefix(line, "*") {

                  size, _ := strconv.Atoi(strings.TrimSpace(strings.Replace(line, "*", "", -1)))

                  if strings.HasPrefix(line_before, "*") {

                      array_size = array_size * size

                  } else if line_before == "" {

                      array_size = size

                  } 

              }

              if array_size > 0 && ! strings.HasPrefix(line, "*") && ! strings.HasPrefix(line, "$") {

                  array_size--

              }

              if line_before != "" && array_size == 0 {

                  _print(false,"stop reading response sending STOP to channel")

                  stop <- true

                  close(stop)

                  return

              }

              line_before = line

        }

    }

}

func getRedisMasterFromResponse(data []string) (string, error) {

    var ip string

    var port string
    
    for _,line := range data {

        line = strings.TrimSpace(line)

        if ! strings.Contains(line, "*") && ! strings.Contains(line, "$")  {

            if ip == "" {

                ip = line

            } else {

                port = line

            }

        }

    }

    port_int, port_err := strconv.Atoi(port)

    if net.ParseIP(ip) != nil && port_err == nil && port_int < 65535 && port_int > 0 {

        return ip + ":" + port, nil

    }

    return "", errors.New("redis master address:port not found")

}

func getSentinelsFromResponse(data []string) []string {

    hosts := []string{}

    var ip_addr_count int

    for idx,line := range data {

        if idx == 0 && strings.Contains(line, "*") {

            var err error

            ip_addr_count, err = strconv.Atoi(strings.Replace(line, "*", "", -1))

            if err != nil {

                return []string{}

            }

        }

        if strings.TrimSpace(line) == "ip" && (idx + 6) < len(data) {

            ip := strings.TrimSpace(data[idx+2])

            port := strings.TrimSpace(data[idx+6])

            port_int,port_err := strconv.Atoi(port)

            if net.ParseIP(ip) != nil && port_err == nil && port_int < 65535 && port_int > 0 {

                hosts = append(hosts,ip + ":" + port )

                ip_addr_count--

            }

        }

    }

    return hosts

}

func isRedisMasterWritable(conn net.Conn) bool {

    _print(false,"isRedisMasterWritable called")

    fmt.Fprintf(conn,  DEL_REDIS_CMD + " " + REDIS_DEL_KEY + "\n")

    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 3)

    defer cancel()

    response := ReadConnection(ctx,conn,stopRedisResponse)

    resp_str := strings.Join(response,"")

    _print(false,"isRedisMasterWritable DEL response",resp_str, conn.RemoteAddr().String())

    if strings.Contains(resp_str, REDIS_READONLY) {

        return false

    } else if strings.Contains(resp_str,":0") {

        return true

    }

    panic("couldn't determine if master is writable, unsupported redis version?")

}
