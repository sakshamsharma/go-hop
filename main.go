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

    // Initialize the listening server
    server := Server{sourceHost, sourcePortNumber, "tcp"}

    // Use piper function to handle connections
    server.Listen(piper)
}

// Establish connection with the client, and begin duplex pipe
func piper(scon net.Conn) {
    ccon, err := net.Dial("tcp", targetHost + ":" + strconv.Itoa(targetPortNumber))
    checkError(err)

    // To quit one side when other has quit as well
    quit := make(chan bool)

    // To figure out when one pipe closes
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

        // Important
        // Helps break the other side's blocked channel
        c1.Close()

        // Let main thread know it can proceed with quits
        wait <- true
    }()

    var len int64
    var err error

    for {
        select {
        case <-quit:
            return
        default:
            // Read up to 1000 bytes and pack them off
            len, err = io.CopyN(c1, c2, 1000)
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
