package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter your username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		_, err = conn.Write([]byte("@join " + username))
		if err != nil {
			fmt.Println("Error joining chat:", err)
			return
		}

		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			return
		}

		response := strings.TrimSpace(string(buffer[:n]))
		if response == "Invalid: Username already exists." {
			fmt.Println("Username already used, please choose another!")
		} else {
			fmt.Println(response)
			break
		}
	}

	fmt.Print("Joined the game! You have one Bulbasaur,open your bag to check!\nUsages:\n" +
		"1.Open your pokedex\n" +
		"2.Catch random 4 Pokemon\n" +
		"3.List the players\n" +
		"4.Invite player to join the battle\n" +
		"5.Quit the game\n" +
		"Enter your choice: ")

	go receiveMessages(conn)

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
		if strings.Split(text, " ")[0] == "@battle" {
			break
		}
	}
}

func receiveMessages(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}
		fmt.Println(string(buffer[:n]))
		if string(buffer[:n]) == "You are out of the game" {
			os.Exit(0)
			break
		}
	}
}
