# Go Telnet/TCP Chat Server [![CircleCI](https://circleci.com/gh/dbnegative/go-telnet-chatserver.svg?style=svg)](https://circleci.com/gh/dbnegative/go-telnet-chatserver)
A basic multi client, multiroom tcp chat server written in GO, can be accessed with telnet. I wrote this to further explore GO routines, unbuffered channels and the net package. Only tested on OSX. Supports basic multi room and multiple simultaneous connections.  

## Build

```
go build server.go
```

## Usage

* Starting the server:
```
/server --help
Usage of ./server:
  -ip string
    	IP address to listen on (default "127.0.0.1")
  -port string
    	Port to listen on (default "8181")

./server
```

* Connecting to the server:

```
telnet 127.0.0.1 8181
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
please enter name: dbnegative
---------------------------
* Welcome dbnegative *
---------------------------
HELP:
---------------------------
\listrooms list all online users
\create create a new room
\join join a room
\help prints all available commands
\quit quit
---------------------------
\create
please enter room name: testroom
---------------------------
* room testroom has been created *
---------------------------
hi
---------------------------
* 29/03/2016 19:18:08 * (dbnegative): "hi" *
---------------------------
```

