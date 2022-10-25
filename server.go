package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	//SERVER
	SERVER_TYPE = "tcp"

	//ERRORS
	ERR_SERVER_FAILURE = "SERVER FAILED TO START"
)

// structures
var usernames = make(map[int]string)
var user_list = make(map[int]*client)
var users = make(map[int]net.Conn)

type client struct {
	conn     net.Conn
	username    string
	num_clients      int
	num_clients_names int
}

// Tcp server
func main() {
	var host string
	var port string
	fmt.Println("Enter host: e.x. localhost");
	fmt.Scan(&host);
	fmt.Println("Enter port: e.x. 8888");
	fmt.Scan(&port);
	connect(host, port);
	fmt.Println("Server Running...")
}
func connect(host string, port string){
	PORT        = port
	SERVER_ADDR = host
	server, err := net.Listen(SERVER_TYPE, SERVER_ADDR+":"+PORT)
	if err != nil {
		fmt.Println(time.Now().Format(time.Stamp)+": Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println(time.Now().Format(time.Stamp) + ": Listening on " + SERVER_ADDR + ":" + PORT)
	fmt.Println(time.Now().Format(time.Stamp) + ": Waiting for client...")

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println(time.Now().Format(time.Stamp)+": Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println(time.Now().Format(time.Stamp) + ": client connected")
		go clientIn(connection)
	}
}

func clientIn(connection net.Conn) {
	var user client
forLoop:
	for {
		var remaining string
		buffer := make([]byte, 1024)

		mLen, err := connection.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			fmt.Println(time.Now().Format(time.Stamp)+": Error reading:", err.Error())
		}
		input := string(buffer[:mLen])
		command, remaining, _ := strings.Cut(input, " ")

		if err != nil && err.Error() == "EOF" {
			command = "/exit\n"
		}
		fmt.Println(time.Now().Format(time.Stamp) + ": user: " + user.username + " entered command " + command)
		fmt.Println(remaining)
		switch command {
		case "/u":	// Sets username
			setUsername(remaining, &user, connection)
		case "/snd": // Broadcasts specified message
			broadcast(&user, remaining)
		case "/psend": // Sends a private message to specified user
			toUser, msg, _ := strings.Cut(remaining, " ")
			toConn := find(toUser)
			if toConn == nil {
				connection.Write([]byte("User not found"))
				break
			}
			send(msg, toConn, &user)
			break

		case "/exit\n": // Disconnects from server
			leave(&user, connection)
			break forLoop
		case "/list\n": // lists all online users
			fmt.Println("command >>  " + command)
			listusers(connection)
		default:
			connection.Write([]byte("Commands:\n  /list - lists online users on server\n /user <username> - sets your unique username\n  /hi <receipient username> <message> - \n /exit - leaves server"))

		}
	}
}

//This function does username validation to ensure unique usernames
func setUsername(name string, user *client, connection net.Conn) {
	var valid = 0
	for i := 0; i < len(usernames); i++ {
		if usernames[i] == name {
			valid = 1
		}
	}
	if valid == 1 {
		connection.Write([]byte("This username is taken please try another one"))
	} else {
		user.conn = connection

		fmt.Println(time.Now().Format(time.Stamp) + ": user: " + name + " connected with " + connection.RemoteAddr().String())
		if user.num_clients_names > 0 {
			connection.Write([]byte("You already have a username"))
		} else {
			usernames[len(usernames)] = name
			user_list[len(user_list)] = user
			users[len(users)] = connection
			user.num_clients = len(usernames)
			fmt.Println(user.num_clients)
			user.username = name
			user.num_clients_names = 1
			fmt.Println(time.Now().Format(time.Stamp) + ": added user :\t" + user.username)
			broadcast(user, user.username+" has joined the server.")

		}
	}

}


// Facilitates the ability to send a message to specified user
func send(msg string, conn net.Conn, sender *client) {
	if sender.username != "" {
		conn.Write([]byte(sender.username + " says: \n" + msg + "\n"))
	}
}

// Facilitates the ability to list all connected users
func listusers(connection net.Conn) {
	for i := 0; i < len(usernames); i++ {
		connection.Write([]byte("*" + usernames[i]))
		fmt.Println(time.Now().Format(time.Stamp) + ": Online users:\t" + strconv.Itoa(len(user_list)))
	}
}

// Allows server to check if specified user exists/ is currently connected
func find(username string) net.Conn {
	for i := 0; i < len(user_list); i++ {
		if user_list[i].username == (username + "\n") {
			return user_list[i].conn
		}
	}
	return nil
}

// Broacasts a message to all clients connected to server
func broadcast(sender *client, msg string) {

	for i := 0; i < len(users); i++ {
		if users[i] != sender.conn {
			send(msg, users[i], sender)
		}
	}
}

//leave server
func leave(user *client, conn net.Conn) {
	delete(usernames, (user.num_clients - 1))
	delete(users, (user.num_clients - 1))
	if len(user.username) > 0 {
		fmt.Println(time.Now().Format(time.Stamp) + ": " + user.username + "has disconnected")
		broadcast(user, user.username+" has left the server.")
	} else {
		fmt.Println(time.Now().Format(time.Stamp) + ": Unknown Client has disconnected") //Need to add here that it broadcasts to all users when someone disconnects
	}
	conn.Close()
}
