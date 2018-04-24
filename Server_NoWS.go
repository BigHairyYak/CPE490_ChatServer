package main

import (
	"net"
	"log"
	"bufio"
	"encoding/json"
	"strings"
	"io"
	"fmt"
)

type Message struct{
	Username	string`json:"username"`
	Command		string`json:"command"`
	Message		string`json"message"`
}

func check(err error){
	if err != nil{
		log.Fatal(err)
	}
}

var connections = map[net.Conn]string{}	//initialize map of connections
var joinMsg Message

func main(){
	listener, err := net.Listen("tcp", "localhost:9000")
	check(err)
	defer listener.Close()

	for {
		conn, _ := listener.Accept()	//Waits for and accepts next connection
		/*print("local: ")
		println(conn.LocalAddr())
		print("remote: ")
		println(conn.RemoteAddr())
		println("Accepted connection")*/

		newReader := *bufio.NewReader(conn) 	//begin reading from connection
		JM, _ := newReader.ReadBytes('}')	//determine if actual client
		err := json.Unmarshal(JM, &joinMsg)
		check(err)

		if joinMsg.Command == "join"{	//if actual client with proper JSON
			connections[conn] = joinMsg.Username//Add new connection to map
			broadcastArrival(joinMsg.Username)	//if actual client, welcome
		} else{
			conn.Close()						//if not, close connection
		}
		defer conn.Close()				//Ignore warning of a memory leak
		go broadcastMessages(newReader)
	}
}

func broadcastArrival(newUser string){
	for connection, user := range connections{
		fmt.Println("Connection: ", connection, "User: ", user)
		send(connection, newUser + " has connected!")
	}
}

func broadcastMessage(message string){
	for connection := range connections{
		send(connection, message)
	}
}

func send(c net.Conn, m string){	//lazy programming
	strings.TrimRight(m,"\n")	//redundant, but force end with \n
	io.WriteString(c, m+"\n")		//otherwise clients read forever
}


func broadcastMessages(reader bufio.Reader) {
	var m Message
	for {
		msg, _ := reader.ReadBytes('}')	//receives JSON, ends with bracket
		err := json.Unmarshal(msg, &m)
		check(err)

		m.Command = strings.Trim(m.Command, "\r") //remove return keys, if any
		switch m.Command {
		case "say":
			broadcastMessage(m.Username + ": " + m.Message)
		case "shout":
			if m.Message == "" {
				broadcastMessage(m.Username + " screams into the void")
			} else {
				broadcastMessage(m.Username + ": " + strings.ToUpper(m.Message))
			}
		case "quit", "disconnect":
			for connection, user := range connections{
				println(user + ", compared to " + m.Username)
				if user == m.Username{				//if user is found
					connection.Close()				//close connection
					delete(connections, connection)	//remove from map
					broadcastMessage(m.Username + " has quit")
					return
				}
			}
		default:
			broadcastMessage(m.Username + ": " + m.Message)
		}
	}
}