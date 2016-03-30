package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type client struct {
	Conn    net.Conn
	Name    string
	Message chan string
	Room    string
}

type chatRoom struct {
	name     string
	messages chan string
	members  map[string]*client
}

var flagIP = flag.String("ip", "127.0.0.1", "IP address to listen on")
var flagPort = flag.String("port", "8181", "Port to listen on")
var customTime = "02/01/2006 15:04:05"
var help = map[string]string{
	"\\quit":      "quit\n",
	"\\listrooms": "list all online users\n",
	"\\create":    "create a new room\n",
	"\\join":      "join a room\n",
	"\\help":      "prints all available commands\n",
}
var roomList = map[string]*chatRoom{}

func main() {
	flag.Parse()

	//start listener
	listener, err := net.Listen("tcp", *flagIP+":"+*flagPort)
	if err != nil {
		log.Fatalf("could not listen on interface %v:%v error: %v ", *flagIP, *flagPort, err)
	}
	defer listener.Close()
	log.Println("listening on: ", listener.Addr())

	//main listen accept loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("could not accept connection %v ", err)
		}
		//create new client on connection
		go createclient(conn)
	}
}

func createclient(conn net.Conn) {

	log.Printf("createclient: remote connection from: %v", conn.RemoteAddr())

	name, err := readInput(conn, "Please Enter Name: ")
	if err != nil {
		panic(err)
	}

	writeFormattedMsg(conn, "Welcome "+name)

	//init client struct
	client := &client{
		Message: make(chan string),
		Conn:    conn,
		Name:    name,
		Room:    "",
	}

	log.Printf("new client created: %v %v", client.Conn.RemoteAddr(), client.Name)

	//spin off seperate send, recieve
	go client.send()
	go client.recieve()

	//print help
	writeFormattedMsg(conn, help)
}

func (c *client) close() {
	c.leave()
	c.Conn.Close()
	c.Message <- "\\quit"
}

func (c *client) recieve() {
	for {
		msg := <-c.Message
		if msg == "\\quit" {
			roomList[c.Room].announce(c.Name + " has left..")
			break
		}
		log.Printf("recieve: client(%v) recvd msg: %s ", c.Conn.RemoteAddr(), msg)
		writeFormattedMsg(c.Conn, msg)
	}
}

func (c *client) send() {
Loop:
	for {
		msg, err := readInput(c.Conn, "")
		if err != nil {
			panic(err)
		}

		if msg == "\\quit" {
			c.close()
			log.Printf("%v has left..", c.Name)
			break Loop
		}

		if c.command(msg) {
			log.Printf("send: msg: %v from: %s", msg, c.Name)
			send := time.Now().Format(customTime) + " * (" + c.Name + "): \"" + msg + "\""

			for _, v := range roomList {
				for k := range v.members {
					if k == c.Conn.RemoteAddr().String() {
						v.messages <- send
					}
				}
			}

		} //validate
	} //for
} //end

func (c *client) command(msg string) bool {
	switch {
	case msg == "\\listrooms":
		c.Conn.Write([]byte("-------------------\n"))
		for k := range roomList {
			count := 0
			for range roomList[k].members {
				count++
			}
			c.Conn.Write([]byte(k + " : online members(" + strconv.Itoa(count) + ")\n"))
		}
		c.Conn.Write([]byte("-------------------\n"))
		return false
	case msg == "\\join":
		c.join()
		return false
	case msg == "\\help":
		writeFormattedMsg(c.Conn, help)
		return false
	case msg == "\\create":
		c.create()
		return false
	}
	return true
}

func (c *client) join() {

	roomName, err := readInput(c.Conn, "Please enter room name: ")
	if err != nil {
		panic(err)
	}

	if cr := roomList[roomName]; cr != nil {
		cr.members[c.Conn.RemoteAddr().String()] = c

		if c.Room != "" {
			c.leave()
			cr.announce(c.Name + " has left..")
		}

		c.Room = roomName
		writeFormattedMsg(c.Conn, c.Name+" has joined "+cr.name)
		cr.announce(c.Name + " has joined!")
	} else {
		writeFormattedMsg(c.Conn, "error: could not join room")
	}
}

//leave current room
func (c *client) leave() {
	//only if room is not empty
	if c.Room != "" {
		delete(roomList[c.Room].members, c.Conn.RemoteAddr().String())
		log.Printf("leave: removing user %v from room %v: current members: %v", c.Name, c.Room, roomList[c.Room].members)
		writeFormattedMsg(c.Conn, "leaving "+c.Room)
	}
}

func (c *client) create() {

	roomName, err := readInput(c.Conn, "Please enter room name: ")
	if err != nil {
		panic(err)
	}
	//if already a member of another room, leave that one first
	if roomName != "" {
		cr := createRoom(roomName)
		cr.members[c.Conn.RemoteAddr().String()] = c

		if c.Room != "" {
			c.leave()
			roomList[c.Room].announce(c.Name + " has left..")
		}
		// set clients room to new room
		c.Room = cr.name
		// add new room to map
		roomList[cr.name] = cr
		cr.announce(c.Name + " has joined!")

		writeFormattedMsg(c.Conn, "* room "+cr.name+" has been created *")
	} else {
		writeFormattedMsg(c.Conn, "* error: could not create room \""+roomName+"\" *")
	}
}

func writeFormattedMsg(conn net.Conn, msg interface{}) error {
	_, err := conn.Write([]byte("---------------------------\n"))
	t := reflect.ValueOf(msg)
	switch t.Kind() {
	case reflect.Map:
		for k, v := range msg.(map[string]string) {
			_, err = conn.Write([]byte(k + " : " + v))
		}
		break
	case reflect.String:
		v := reflect.ValueOf(msg).String()
		_, err = conn.Write([]byte(v + "\n"))
		break
	} //switch
	conn.Write([]byte("---------------------------\n"))

	if err != nil {
		return err
	}
	return nil //todo
}

func readInput(conn net.Conn, qst string) (string, error) {
	conn.Write([]byte(qst))
	s, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("readinput: could not read input from stdin: %v from client %v", err, conn.RemoteAddr().String())
		return "", err
	}
	s = strings.Trim(s, "\r\n")
	return s, nil
}

func createRoom(name string) *chatRoom {
	c := &chatRoom{
		name:     name,
		messages: make(chan string),
		members:  make(map[string]*client, 0),
	}
	log.Printf("creating room %v", c.name)
	//spin off new routine to listen for messages
	go func(c *chatRoom) {
		for {
			out := <-c.messages
			if out == "\\kill" {
				log.Printf("chatroom: killing \"%v\"", c.name)
				//remove from room map
				delete(roomList, c.name)
				//kill routine - only if no members
				break
			}
			for _, v := range c.members {
				v.Message <- out
				log.Printf("createroom: broadcasting msg in room: %v to member: %v", c.name, v.Name)
			}
		}
	}(c)

	//poll member list and clean up after members = 0
	go func(c *chatRoom) {
		for {
			if len(c.members) == 0 {
				log.Printf("chatroom: zero members cleaning chatroom \" %v \"", c.name)
				c.messages <- "\\kill"
				break
			}
		}
	}(c)
	return c
}

func (c *chatRoom) announce(msg string) {
	c.messages <- "* " + msg + " *"
}
