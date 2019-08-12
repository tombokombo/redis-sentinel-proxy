package main

import (
    "context"
    "net"
    "bufio"
    "fmt"
    "strings"
    "time"
    "io"
)

func _print(just_verbose bool, str ...string) {

    if ! cfg_debug && ! just_verbose {

        return

    }

    if cfg_verbose {

        fmt.Println(strings.Join(str, " "))

    }

}

func check_warn(err error, str ...string){

  if err != nil {

      fmt.Println(strings.Join(str, " "),err)

  }

}

func check_strict(err error) {
    //stolen somewhere
    if err != nil {

        panic(err)

    }

}

func check_warn_bool(err error, str ...string) bool {
    //stolen
    if err != nil {

        fmt.Println(strings.Join(str, " "),err)

        return true

    }

    return false

}

func dataCopyToClient(from net.Conn, to *Connection) {
    //stolen from redis-ellison
    defer from.Close()

    io.Copy(from, to.conn)

}

func dataCopyFromClient(from *Connection, to net.Conn) {
    //stolen from redis-ellison
    defer from.Close()

    io.Copy(from.conn, to)

}

func getFirstConnFromArr(hosts []string, ch chan net.Conn) {

    for _,host := range hosts {

      go func(h string,ch chan net.Conn) {

        conn, err := net.DialTimeout("tcp", h , time.Duration(3 * time.Second))

        if err == nil {

            select {

                case ch <-conn:

                default:

                    conn.Close()

            }

        }

      }(host, ch)

  }

}

func isConnAlive(conn net.Conn) bool{

    buf := []byte("conntest")

    n, err := conn.Write(buf)

    if err != nil {

        if strings.Contains(err.Error(),"use of closed network connection") {

            return false

        }

    }

    if n == len(buf) {

          b := make([]byte,1024)

          n,err := conn.Read(b)

          if err != nil {

              return false

          }

          if n > 0 {

              return true

          }

    }

    return false
}

func ReadConnection(ctx context.Context,conn net.Conn,stop_fn func(chan bool,chan string)) []string {

    msg_ch := make(chan string)

    stop_ch := make(chan bool)

    stop_data_ch := make(chan string)

    go stop_fn(stop_ch,stop_data_ch)

    go func(cn net.Conn,msg_ch chan string,stop_data_ch chan string) {

        defer close(msg_ch)

        defer close(stop_data_ch)

        scanner := bufio.NewScanner(cn)

        for {

            gotdata := scanner.Scan()

            if gotdata {

               txt:= scanner.Text()

               msg_ch <- txt

               stop_data_ch <- txt

            } else {

                _print(false,"readConnection scan end, no more data")

                break

            }

        }

    }(conn, msg_ch,stop_data_ch)

    data := []string{}

    forloop:
    for {

        select {

            case text := <-msg_ch:

                data = append(data,text)

            case <-ctx.Done():

                _print(false,"ReadConnection ctx done...firing deadline NOW")

                err := conn.SetDeadline(time.Now())

                check_warn(err,"unable to set deadline 1")

                break forloop

            case <- stop_ch:

                _print(false,"ReadConnection STOP channel recived...firing deadline NOW")

                err := conn.SetDeadline(time.Now())

                check_warn(err,"unable to set deadline 2")

                break forloop

        }

    }

    _print(false,"readConnection finished...disabling deadline")

    //sleep needed, because racecond. is shit 
    time.Sleep(10*time.Millisecond)

    err := conn.SetDeadline(time.Time{})

    check_warn(err,"unable to disable deadline")

    return data

}
