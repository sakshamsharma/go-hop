package main

import (
    "io"
    "net"
    "os"
    "strconv"
)

var (
    targetHost string
    targetPortNumber int
)

func main() {
    if len(os.Args) < 5 {
        Error.Fatal("Did not provide hostname / port")
    }

    sourceHost := os.Args[1]
    sourcePortNumber, err := strconv.Atoi(os.Args[2])
    checkError(err)

    targetHost = os.Args[3]
    targetPortNumber, err = strconv.Atoi(os.Args[4])
    checkError(err)

    server := Server{sourceHost, sourcePortNumber, "tcp"}
    server.Listen(piper)
}

func piper(scon net.Conn) {
    ccon, err := net.Dial("tcp", targetHost + ":" + strconv.Itoa(targetPortNumber))
    checkError(err)

    quit := make(chan bool)
    wait := make(chan bool)

    go checkedPipe(ccon, scon, quit, wait, "c2s")
    go checkedPipe(scon, ccon, quit, wait, "s2c")

    <- wait
    quit <- true
    close(quit)
}

func checkedPipe(c1, c2 net.Conn, quit chan bool, wait chan bool, name string) {
    defer func() {
        Info.Println(name, ": Exiting")
        c1.Close()
        wait <- true
    }()

    var len int64
    var err error

    for {
        select {
        case <-quit:
            return
        default:
            len, err = io.CopyN(c1, c2, 200)
            if err != nil || len == 0 {
                return
            }
        }
    }
}

func checkError(err error) {
    if err != nil {
        Error.Fatal(err.Error())
    }
}
