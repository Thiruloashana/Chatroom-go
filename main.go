package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket" //importing the gorilla toolkit which has the the websocket package
)

type Client struct {
    conn *websocket.Conn //conn is of type websocket
	//This field holds a pointer to a websocket connection that is established between the client and the server.
    username string //used to hold the username of the joining client
}

var clients []*Client  //creating an array of pointers to the client struct

func main() {
    http.HandleFunc("/ws", handleWebSocket) //handles the requests from the websocket
    http.HandleFunc("/", handleHome) //handles the requests from http home page
    fmt.Println("Chat room server started.") //prints in the terminal
    log.Fatal(http.ListenAndServe(":9000", nil))  // 9000 is the port used for web development mostly
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "index.html") //linking the html page
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }

    conn, err := upgrader.Upgrade(w, r, nil) //if there are any errors in reading or writing, then err will not be nil
    if err != nil {
        log.Println(err)
        return
    }

    client := &Client{conn: conn}

    clients = append(clients, client)

    fmt.Println("New client connected")

    sendUserPrompt(client) //asking the user to enter the username

    //the loop is used to enter the inputs from the text widgets
    for {
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            removeClient(client)
            return
        }

        if string(msg) == "/quit" {
            removeClient(client)
            sendLeftMessage(client)
            return
        }

        if client.username == "" {
            client.username = string(msg)
            sendJoinedMessage(client)
            continue
        }

        broadcast(client.username+": "+string(msg), msgType)
    }
}

func sendUserPrompt(client *Client) {
    prompt := "Enter your username:"

    err := client.conn.WriteMessage(websocket.TextMessage, []byte(prompt))
    if err != nil {
        log.Println(err)
        removeClient(client)
        return
    }
}

func sendJoinedMessage(client *Client) {
    message := client.username + " joined the chat"
    broadcast(message, websocket.TextMessage)
}

func sendLeftMessage(client *Client) {
    message := client.username + " left the chat"
    broadcast(message, websocket.TextMessage)
}

func broadcast(msg string, msgType int) {
    for _, client := range clients {
        err := client.conn.WriteMessage(msgType, []byte(msg))
        if err != nil {
            log.Println(err)
            removeClient(client)
        }
    }
}

func removeClient(client *Client) {
    for i, c := range clients {
        if c == client {
            clients = append(clients[:i], clients[i+1:]...)
            break
        }
    }
}