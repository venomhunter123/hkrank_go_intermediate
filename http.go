package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

const portSuffix = ":3333"
var address = "http://127.0.0.1" + portSuffix

type Lake struct {
    Id   string `json:"id"`
    Name string `json:"name"`
    Area int32  `json:"area"`
}

type Action struct {
    Type    string `json:"type"`
    Payload string `json:"payload"`
}

var store = map[string]Lake{}

func postHandler(w http.ResponseWriter, req *http.Request) {
    var lake Lake
    err := json.NewDecoder(req.Body).Decode(&lake)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    store[lake.Id] = lake

    // Respond with HTTP status 200 OK
    w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, req *http.Request) {
    id := req.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "Missing id parameter", http.StatusBadRequest)
        return
    }

    delete(store, id)

    // Respond with HTTP status 200 OK
    w.WriteHeader(http.StatusOK)
}

func getHandler(w http.ResponseWriter, req *http.Request) {
    id := req.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "Missing id parameter", http.StatusBadRequest)
        return
    }

    lake, ok := store[id]
    if !ok {
        http.Error(w, "Lake not found", http.StatusNotFound)
        return
    }

    // Encode the lake object as JSON and send it in the response
    jsonBytes, err := json.Marshal(lake)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonBytes)
}

func main() {
    reader := bufio.NewReaderSize(os.Stdin, 16*1024*1024)

    stdout, err := os.Create(os.Getenv("OUTPUT_PATH"))
    checkError(err)
    defer stdout.Close()

    writer := bufio.NewWriterSize(stdout, 16*1024*1024)

    actionsCountStr := readLine(reader)
    actionsCount, err := strconv.Atoi(actionsCountStr)
    checkError(err)

    // Start HTTP server to handle requests
    http.HandleFunc("/get", getHandler)
    http.HandleFunc("/post", postHandler)
    http.HandleFunc("/delete", deleteHandler)
    go http.ListenAndServe(portSuffix, nil)
    time.Sleep(100 * time.Millisecond)

    var actions []string

    // Read actions from input
    for i := 0; i < actionsCount; i++ {
        actionsItem := readLine(reader)
        actions = append(actions, actionsItem)
    }

    // Process each action
    for _, actionStr := range actions {
        var action Action
        err := json.Unmarshal([]byte(actionStr), &action)
        checkError(err)

        switch action.Type {
        case "post":
            // Perform HTTP POST request
            resp, err := http.Post(address+"/post", "application/json", strings.NewReader(action.Payload))
            checkError(err)
            resp.Body.Close()
        case "delete":
            // Perform HTTP DELETE request
            client := &http.Client{}
            req, err := http.NewRequest("DELETE", address+"/delete?id="+action.Payload, nil)
            checkError(err)
            resp, err := client.Do(req)
            checkError(err)
            resp.Body.Close()
            if resp.StatusCode != http.StatusOK {
                fmt.Fprintf(writer, "%s\n", resp.Status)
                continue
            }
        case "get":
            // Perform HTTP GET request
            resp, err := http.Get(address + "/get?id=" + action.Payload)
            checkError(err)
            defer resp.Body.Close()
            if resp.StatusCode != http.StatusOK {
                fmt.Fprintf(writer, "%s\n", resp.Status)
                continue
            }
            var lake Lake
            err = json.NewDecoder(resp.Body).Decode(&lake)
            checkError(err)
            fmt.Fprintf(writer, "%s\n", lake.Name)
            fmt.Fprintf(writer, "%d\n", lake.Area)
        }
    }

    fmt.Fprintf(writer, "\n")
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
