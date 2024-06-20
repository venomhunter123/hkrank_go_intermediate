package main

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "os"
    "strconv"
    "strings"
)

const (
    maxBufferSize = 1024
    address       = "127.0.0.1:3333"
)

/*
 * Complete the 'TCPServer' function below.
 *
 * The function accepts chan bool ready as a parameter.
 */

func TCPServer(ready chan bool) {
    // Resolve TCP address
    tcpAddr, err := net.ResolveTCPAddr("tcp", address)
    if err != nil {
        panic(err)
    }

    // Create listener
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        panic(err)
    }

    // Signal readiness
    ready <- true

    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    defer conn.Close()

    // Read client message
    buf := make([]byte, maxBufferSize)
    n, err := conn.Read(buf)
    if err != nil {
        return
    }

    // Reverse message
    msg := string(buf[:n])
    reversedMsg := reverseString(msg)

    // Send reversed message back to client
    _, err = conn.Write([]byte(reversedMsg))
    if err != nil {
        return
    }
}

func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

func main() {
    stdout, err := os.Create(os.Getenv("OUTPUT_PATH"))
    checkError(err)
    defer stdout.Close()

    reader := bufio.NewReaderSize(os.Stdin, 16*1024*1024)
    writer := bufio.NewWriterSize(stdout, 16*1024*1024)

    messagesCountStr := readLine(reader)
    messagesCount, err := strconv.Atoi(messagesCountStr)
    checkError(err)

    var messages []string
    for i := 0; i < messagesCount; i++ {
        messagesItem := readLine(reader)
        messages = append(messages, messagesItem)
    }

    ready := make(chan bool)
    go TCPServer(ready)
    <-ready

    reversed, err := tcpClient(messages)
    checkError(err)

    for _, msg := range reversed {
        fmt.Fprintf(writer, "%s\n", msg)
    }
    writer.Flush()
}

func readLine(reader *bufio.Reader) string {
    str, _, err := reader.ReadLine()
    if err == io.EOF {
        return ""
    }

    return strings.TrimRight(string(str), "\r\n")
}

func checkError(err error) {
    if err != nil {
        panic(err)
    }
}

func tcpClient(messages []string) ([]string, error) {
    tcpAddr, err := net.ResolveTCPAddr("tcp", address)
    if err != nil {
        return nil, err
    }

    reversed := []string{}

    for _, msg := range messages {
        conn, err := net.DialTCP("tcp", nil, tcpAddr)
        if err != nil {
            return nil, err
        }
        defer conn.Close()

        _, err = conn.Write([]byte(msg))
        if err != nil {
            return nil, err
        }

        reply := make([]byte, maxBufferSize)
        n, err := conn.Read(reply)
        if err != nil {
            return nil, err
        }

        reversed = append(reversed, string(reply[:n]))
    }

    return reversed, nil
}
