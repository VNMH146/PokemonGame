package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	Name            string
	Addr            *net.UDPAddr
	userCurrentPoke Pokedex
	userPokedex     []Pokedex
	currentPoke     Pokedex
	battlePoke      []Pokedex
}
type Battle struct {
	Player1      *Client
	Player2      *Client
	CurrentPoke1 *Pokedex
	CurrentPoke2 *Pokedex
	CurrentTurn  *net.UDPAddr
}
type Pokedex struct {
	Id       string `json:"ID"`
	Name     string `json:"Name"`
	Level    int    `json:"Level"`
	Exp      int
	Types    []string `json:"types"`
	Link     string   `json:"URL"`
	PokeInfo PokeInfo `json:"Poke-Information"`
}
type PokeInfo struct {
	Hp          int     `json:"HP"`
	Atk         int     `json:"ATK"`
	Def         int     `json:"DEF"`
	SpAtk       int     `json:"Sp.Atk"`
	SpDef       int     `json:"Sp.Def"`
	Speed       int     `json:"Speed"`
	TypeDefense TypeDef `json:"Type-Defenses"`
}
type TypeDef struct {
	Normal   float32
	Fire     float32
	Water    float32
	Electric float32
	Grass    float32
	Ice      float32
	Fighting float32
	Poison   float32
	Ground   float32
	Flying   float32
	Psychic  float32
	Bug      float32
	Rock     float32
	Ghost    float32
	Dragon   float32
	Dark     float32
	Steel    float32
	Fairy    float32
}

var (
	clients    = make(map[string]*Client)
	pokedex    []Pokedex
	battles    = make(map[string]*Client)
	mu         sync.Mutex
	invitation = make(map[string]string)
	games      = make(map[string]*Battle)
	state      *net.UDPAddr
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Server is running on port 8080")

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}
		message := string(buffer[:n])
		handleMessage(message, addr, conn)
	}
}

func handleMessage(message string, addr *net.UDPAddr, conn *net.UDPConn) {
	mu.Lock()
	defer mu.Unlock()

	parts := strings.Split(message, " ")
	command := parts[0]
	senderName := getUsernameByAddr(addr)

	client := clients[senderName]

	switch command {
	case "@join":
		if len(parts) < 2 {
			sendMessageToClient("Invalid: Please provide a username.", addr, conn)
			return
		}

		username := parts[1]
		if checkExist(username) {
			sendMessageToClient("Invalid: Username already exists.", addr, conn)
			return
		}

		clients[username] = &Client{Name: username, Addr: addr}

		// Kiểm tra xem tệp JSON lưu trữ Pokémon của người dùng có tồn tại không
		filePath := username + "_Pokedex.json"
		if _, err := os.Stat(filePath); err == nil {
			// Nếu tệp tồn tại, tải dữ liệu từ tệp
			var savedPokedex []Pokedex
			OpenFile(filePath, &savedPokedex)
			clients[username].userPokedex = savedPokedex
			if len(savedPokedex) > 0 {
				clients[username].userCurrentPoke = savedPokedex[0]
			}
			fmt.Printf("User [%s] reloaded with saved data.\n", username)
		} else {
			// Nếu tệp không tồn tại, khởi tạo người dùng với một Pokémon mặc định
			OpenFile("data/pokedex.json", &pokedex)
			for _, poke := range pokedex {
				if poke.Id == "#0001" {
					clients[username].userCurrentPoke = poke
					clients[username].userCurrentPoke.Level = 1
					clients[username].userPokedex = append(clients[username].userPokedex, poke)
					break
				}
			}
			fmt.Printf("New user [%s] initialized with default Pokemon.\n", username)

			// Lưu tệp JSON cho người dùng mới
			CreateFile(filePath, clients[username].userPokedex)
		}

		sendMessageToClient("["+username+"] Welcome to the POKEMON game!", addr, conn)

	case "5":
		username := getUsernameByAddr(addr)
		delete(clients, username)
		fmt.Print("Player [" + username + "] out the game\n")
		sendMessageToClient("You are out the game", addr, conn)

		if parts[1] == "1" {
			OpenFile("data\\pokedex.json", &pokedex)
			for _, poke := range pokedex {
				if poke.Id == "#0001" {
					client.userCurrentPoke = poke
					client.userCurrentPoke.Level = 1
					client.userPokedex = append(client.userPokedex, client.userCurrentPoke)
				}
			}
			sendMessageToClient("Valid", addr, conn)
		} else if parts[1] == "2" {
			OpenFile("data\\pokedex.json", &pokedex)
			for _, poke := range pokedex {
				if poke.Id == "#0004" {
					client.userCurrentPoke = poke
					client.userCurrentPoke.Level = 1
					client.userPokedex = append(client.userPokedex, client.userCurrentPoke)
				}
			}
			sendMessageToClient("Valid", addr, conn)
		} else if parts[1] == "3" {
			OpenFile("data\\pokedex.json", &pokedex)
			for _, poke := range pokedex {
				if poke.Id == "#0007" {
					client.userCurrentPoke = poke
					client.userCurrentPoke.Level = 1
					client.userPokedex = append(client.userPokedex, client.userCurrentPoke)
				}
			}
			sendMessageToClient("Valid", addr, conn)
		} else {
			sendMessageToClient("Cannot", addr, conn)
		}
		CreateFile(senderName+"_Pokedex.json", client.userPokedex)

		if len(parts) != 2 {
			sendMessageToClient("Invalid command! Please try again!\n", addr, conn)
		} else {
			OpenFile("data\\pokedex.json", &pokedex)
			for _, poke := range pokedex {
				if poke.Name == parts[1] {
					msg := fmt.Sprintf("ID: %s - Name: %s - HP: %d - ATK: %d - DEF: %d - SPEED: %d\n",
						poke.Id, poke.Name, poke.PokeInfo.Hp, poke.PokeInfo.Atk, poke.PokeInfo.Def, poke.PokeInfo.Speed)

					sendMessageToClient(msg, addr, conn)
				}
			}

		}
	case "2":
		if client == nil {
			sendMessageToClient("Error: You must join the game first.", addr, conn)
			return
		}

		// Đảm bảo `userCurrentPoke` đã được khởi tạo
		if client.userCurrentPoke.Id == "" {
			sendMessageToClient("Error: No current Pokémon available. Please start the game properly.", addr, conn)
			return
		}

		// Kiểm tra dữ liệu `pokedex`
		if len(pokedex) == 0 {
			OpenFile("data/pokedex.json", &pokedex)
		}
		getPoke := RollPoke(client.userCurrentPoke)
		ListPokemon := "Your new pokemon:\n"
		for _, poke := range getPoke {
			ListPokemon += fmt.Sprintf("[ID: %s --Name: %s -- Level: %d]\n", poke.Id, poke.Name, poke.Level)
		}
		sendMessageToClient(ListPokemon, addr, conn)
		client.userPokedex = append(client.userPokedex, getPoke...)
		CreateFile(senderName+"_Pokedex.json", client.userPokedex)
	case "1":
		if client == nil {
			sendMessageToClient("Error: You must join the game first.", addr, conn)
			return
		}
		msg := "Your Bag:\n"
		for _, poke := range client.userPokedex {
			msg += fmt.Sprintf("ID: %s - Name: %s [Level: %d] - HP: %d - ATK: %d - DEF: %d - SPEED: %d\n",
				poke.Id, poke.Name, poke.Level, poke.PokeInfo.Hp, poke.PokeInfo.Atk, poke.PokeInfo.Def, poke.PokeInfo.Speed)
		}
		sendMessageToClient(msg, addr, conn)
	case "p":
		if len(parts) != 4 {
			sendMessageToClient("Invalid input! Please try again!\n", addr, conn)
		} else {
			confirm := "Your pokemon choosen:\n"
			if checkPokeExist(parts[1], parts[2], parts[3], client) {
				for _, poke := range client.userPokedex {
					if parts[1] == poke.Id {
						confirm += poke.Name + " "
						client.battlePoke = append(client.battlePoke, poke)
					}
					if parts[2] == poke.Id {
						confirm += poke.Name + " "
						client.battlePoke = append(client.battlePoke, poke)
					}
					if parts[3] == poke.Id {
						confirm += poke.Name + " "
						client.battlePoke = append(client.battlePoke, poke)
					}
				}
				confirm += "\n(Usage: Enter start to start battle!)\n"
				sendMessageToClient(confirm, addr, conn)
			} else {
				sendMessageToClient("Poke you choose is not have in your pokedex!", addr, conn)
			}
		}
	case "3":
		competitors := "Current player:\n"
		for _, user := range clients {
			if user.Name != senderName {
				competitors += fmt.Sprintf("[Player: %s]\n", user.Name)
			}
		}
		sendMessageToClient(competitors, addr, conn)
	case "4":
		for _, user := range clients {
			if user.Name == parts[1] || parts[1] != senderName { // if exist username like this then
				for _, bt := range battles {
					if user == bt {
						sendMessageToClient(parts[1]+" is in battle, please try later!", addr, conn)
						return
					}
				}
				invitation[addr.String()] = senderName
				invitation[user.Addr.String()] = parts[1] // invitation with index string of that user addr ---> get value of part[1]

			} else if parts[1] == senderName {
				sendMessageToClient("Cannot invite yourself!!!", addr, conn)
				return
			}
		}
		sendMessageToClient("Waiting for your competitor!", addr, conn)
		sendPrivateMessage(senderName+" send you a request to battle!(accept yes/no)\n", parts[1], conn, addr)
		// I think error happen when 2 clients @invite will add into battle
	case "accept":
		if strings.ToLower(parts[1]) == "yes" {
			receiverName := invitation[addr.String()]
			var inviterName string
			for _, invite := range invitation {
				if receiverName != invite {
					inviterName = invite
				}
			}
			for _, user := range clients {
				if user.Name == inviterName {
					sendMessageToClient(senderName+" has accepted the battle\n(Usage: @pick #id_pokemon1 #id_pokemon2 #id_pokemon3)\n", user.Addr, conn)
					battles[user.Addr.String()] = user // inviter client

				}
				if user.Addr.String() == addr.String() {
					battles[addr.String()] = user // receiver client
				}
			}
			sendMessageToClient("You are join the battle!\n(Usage: @pick #id_pokemon1 #id_pokemon2 #id_pokemon3)\n", addr, conn)
		} else if strings.ToLower(parts[1]) == "no" {
			var inviterName string
			for _, invite := range invitation {
				if senderName != invite {
					inviterName = invite
				}
			}
			for _, user := range clients {
				if user.Name == inviterName {
					sendMessageToClient("Your competitor is decline\nChoose another user or other task", user.Addr, conn)
				}
			}
			delete(invitation, addr.String())
			sendMessageToClient("You decline successfull\nLet continue other tasks\n", addr, conn)
		} else {
			sendMessageToClient("Invalid command!\n", addr, conn)
		}
	case "start":
		player, inBattle := battles[addr.String()]
		if !inBattle {
			sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
			return
		}
		gameKey := ""
		var opponent *Client
		for _, bat := range battles {
			if bat.Addr.String() != addr.String() {
				opponent = battles[bat.Addr.String()]
				gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
			}
		}
		if _, inBattle := games[gameKey]; inBattle {
			sendMessageToClient("A game is already in process.", addr, conn)
			return
		}
		if opponent.battlePoke[0].PokeInfo.Speed > player.battlePoke[0].PokeInfo.Speed {
			state = opponent.Addr
		} else {
			state = player.Addr
		}
		sendMessageToClient("You first", state, conn)

		battle := &Battle{
			Player1:      player,
			Player2:      opponent,
			CurrentPoke1: &player.battlePoke[0],
			CurrentPoke2: &opponent.battlePoke[0],
			CurrentTurn:  state,
		}
		games[gameKey] = battle
	case "attack":
		handleAttack(addr, conn)
	case "switch":
		if len(parts) != 2 {
			sendMessageToClient("Invalid command!", addr, conn)
			return
		}
		handleSwitch(conn, addr, parts[1])
	case "surrender":
		opponent, inBattle := battles[addr.String()]

		if !inBattle {
			sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
			return
		}
		var player *Client
		gameKey := ""
		for _, bat := range battles {
			if bat.Addr.String() != addr.String() {
				player = battles[bat.Addr.String()]
				gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
			}
		}
		game, inGame := games[gameKey]
		if !inGame {
			sendMessageToClient("You are not already\n(Usage: @play to ready the battle)", addr, conn)
			return
		}

		winner := player
		loser := opponent
		// Phân phối kinh nghiệm
		distributeExp(winner, loser)
		conn.WriteToUDP([]byte(fmt.Sprintf("Game over! %s wins!", winner.Name)), winner.Addr)
		conn.WriteToUDP([]byte(fmt.Sprint("Game over! You lose!", loser.Name)), loser.Addr)

		delete(games, gameKey)
		delete(battles, game.Player1.Addr.String())
		delete(battles, game.Player2.Addr.String())
	default:
		sendMessageToClient("Invalid command", addr, conn)
	}

}

func getUsernameByAddr(addr *net.UDPAddr) string {
	for _, client := range clients {
		if client.Addr.IP.Equal(addr.IP) && client.Addr.Port == addr.Port {
			return client.Name
		}
	}
	return ""
}

func sendMessageToClient(message string, addr *net.UDPAddr, conn *net.UDPConn) {
	maxMessageSize := 512 // Giới hạn kích thước mỗi gói tin (thấp hơn giới hạn UDP để an toàn)

	// Chia nhỏ thông báo nếu cần
	for len(message) > 0 {
		if len(message) > maxMessageSize {
			conn.WriteToUDP([]byte(message[:maxMessageSize]), addr)
			message = message[maxMessageSize:]
		} else {
			_, err := conn.WriteToUDP([]byte(message), addr)
			if err != nil {
				fmt.Println("Error sending message:", err)
			}
			break
		}
	}
}

func checkExist(name string) bool {
	_, exist := clients[name]
	if exist {
		return true
	} else {
		return false
	}
}

func sendPrivateMessage(message, recipient string, conn *net.UDPConn, addr *net.UDPAddr) {
	client, exists := clients[recipient]
	if !exists {
		fmt.Println("Recipient not found:", recipient)
		conn.WriteToUDP([]byte("NotFound"), addr)
		return
	}
	conn.WriteToUDP([]byte(message), client.Addr)
}
func OpenFile(fileName string, key interface{}) {
	inFile, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Error to open file: ", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(inFile)
	err = decoder.Decode(key)
	if err != nil {
		log.Fatal("Error decoding file: ", err)
		os.Exit(1)
	}
	inFile.Close()

	// Làm sạch dữ liệu (ví dụ: xóa ký tự xuống dòng trong tên Pokémon)
	if pokedexData, ok := key.(*[]Pokedex); ok {
		for i := range *pokedexData {
			(*pokedexData)[i].Name = strings.ReplaceAll((*pokedexData)[i].Name, "\n", "")
		}
	}
}

func CreateFile(fileName string, key interface{}) {
	jsonData, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON: ", err)
		return
	}
	err = ioutil.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file: ", err)
		return
	}
	fmt.Println(fileName + " updated!")
}
func RollPoke(userCurrentPoke Pokedex) []Pokedex {
	var userPokedex []Pokedex
	for i := 0; i < 4; i++ {
		getId := rand.Intn(1025-1) + 1
		var Idpoke string
		if getId < 10 {
			Idpoke = "#000" + strconv.Itoa(getId)
		} else if getId >= 10 && getId < 100 {
			Idpoke = "#00" + strconv.Itoa(getId)
		} else if getId >= 100 && getId < 1000 {
			Idpoke = "#0" + strconv.Itoa(getId)
		} else {
			Idpoke = "#" + strconv.Itoa(getId)
		}
		for _, poke := range pokedex {
			if Idpoke == poke.Id {
				userCurrentPoke = poke
				userCurrentPoke.Level = 1
				userPokedex = append(userPokedex, userCurrentPoke)
			}
		}
	}
	return userPokedex
}
func checkPokeExist(poke1, poke2, poke3 string, client *Client) bool {
	var allExist = false
	epoke1, epoke2, epoke3 := false, false, false
	for _, poke := range client.userPokedex {
		if poke1 == poke.Id {
			epoke1 = true
		}
		if poke2 == poke.Id {
			epoke2 = true
		}
		if poke3 == poke.Id {
			epoke3 = true
		}
	}
	if epoke1 && epoke2 && epoke3 {
		allExist = true
	}
	return allExist
}
func handleAttack(addr *net.UDPAddr, conn *net.UDPConn) {
	opponent, inBattle := battles[addr.String()]
	if !inBattle {
		sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
		return
	}

	gameKey := ""
	var player *Client
	for _, bat := range battles {
		if bat.Addr.String() != addr.String() {
			player = battles[bat.Addr.String()]
			gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
		}
	}

	game, inGame := games[gameKey]
	if !inGame {
		sendMessageToClient("Game not in progress.", addr, conn)
		return
	}

	if addr.String() != state.String() {
		sendMessageToClient("Not your turn!", addr, conn)
		return
	}

	// Xác định attacker và defender
	var attacker, defender *Pokedex
	if addr.String() == game.Player1.Addr.String() {
		attacker = game.CurrentPoke1
		defender = game.CurrentPoke2
	} else {
		attacker = game.CurrentPoke2
		defender = game.CurrentPoke1
	}

	// Kiểm tra nếu Pokémon tấn công đã bị hạ
	// Kiểm tra nếu Pokémon tấn công hoặc phòng thủ đã bị hạ
	if attacker.PokeInfo.Hp <= 0 {
		sendMessageToClient("Your current Pokémon has fainted! Please switch to another Pokémon.", addr, conn)
		return
	}

	if defender.PokeInfo.Hp <= 0 {
		sendMessageToClient(fmt.Sprintf("Your opponent's Pokémon has fainted! Waiting for them to switch Pokémon."), addr, conn)
		return
	}

	// Random chọn kiểu tấn công
	attackType := rand.Intn(2)
	var damage int
	if attackType == 0 { // Normal attack
		damage, _ = getDmgNumber(*attacker, *defender)
	} else { // Special attack
		_, damage = getDmgNumber(*attacker, *defender)
	}

	// Log chi tiết trước khi trừ HP
	fmt.Printf("[LOG] %s (HP: %d) attacks %s (HP: %d) with %s attack.\n",
		attacker.Name, attacker.PokeInfo.Hp, defender.Name, defender.PokeInfo.Hp,
		map[int]string{0: "Normal", 1: "Special"}[attackType])

	// Trừ HP của defender
	defender.PokeInfo.Hp -= damage
	if defender.PokeInfo.Hp < 0 {
		defender.PokeInfo.Hp = 0
	}

	// Log chi tiết sau khi trừ HP
	fmt.Printf("[LOG] %s dealt %d damage to %s. Remaining HP: %d\n",
		attacker.Name, damage, defender.Name, defender.PokeInfo.Hp)

	// Gửi thông báo trạng thái
	sendMessageToClient(fmt.Sprintf("%s attacked %s! Your %s's HP: %d\n%s's opponent - HP: %d",
		attacker.Name, defender.Name, attacker.Name, attacker.PokeInfo.Hp, defender.Name, defender.PokeInfo.Hp), opponent.Addr, conn)

	sendMessageToClient(fmt.Sprintf("%s attacked and remaining HP: %d", attacker.Name, attacker.PokeInfo.Hp), player.Addr, conn)

	// Kiểm tra nếu Pokémon bị hạ
	if defender.PokeInfo.Hp == 0 {
		fmt.Printf("[LOG] %s has fainted.\n", defender.Name)

		if addr.String() == game.Player1.Addr.String() {
			sendMessageToClient(fmt.Sprintf("%s has fainted! Please switch your Pokémon.", game.CurrentPoke2.Name), game.Player2.Addr, conn)
			handlePokemonDefeated(game, conn, game.Player2.Addr)
		} else {
			sendMessageToClient(fmt.Sprintf("%s has fainted! Please switch your Pokémon.", game.CurrentPoke1.Name), game.Player1.Addr, conn)
			handlePokemonDefeated(game, conn, game.Player1.Addr)
		}
		return
	}

	// Đổi lượt chơi
	if addr.String() == game.Player1.Addr.String() {
		state = game.Player2.Addr
	} else {
		state = game.Player1.Addr
	}

	fmt.Printf("[LOG] Turn switched to %s.\n", getUsernameByAddr(state))
}

func handlePokemonDefeated(game *Battle, conn *net.UDPConn, addr *net.UDPAddr) {
	if addr.String() == game.Player1.Addr.String() {
		if len(game.Player1.battlePoke) > 1 {
			sendMessageToClient("Your Pokémon has fainted! Please switch to another Pokémon using @switch <PokemonID>.", game.Player1.Addr, conn)
		} else {
			sendMessageToClient("Game over! You lose!", game.Player1.Addr, conn)
			sendMessageToClient("Game over! You win!", game.Player2.Addr, conn)
			cleanUpGame(game)
		}
	} else if addr.String() == game.Player2.Addr.String() {
		if len(game.Player2.battlePoke) > 1 {
			sendMessageToClient("Your Pokémon has fainted! Please switch to another Pokémon using @switch <PokemonID>.", game.Player2.Addr, conn)
		} else {
			sendMessageToClient("Game over! You lose!", game.Player2.Addr, conn)
			sendMessageToClient("Game over! You win!", game.Player1.Addr, conn)
			cleanUpGame(game)
		}
	}
}

func cleanUpGame(game *Battle) {
	// Xóa game khỏi danh sách
	gameKey := fmt.Sprintf("%s:%s", game.Player1.Name, game.Player2.Name)
	delete(games, gameKey)

	// Xóa thông tin người chơi khỏi battles
	delete(battles, game.Player1.Addr.String())
	delete(battles, game.Player2.Addr.String())

	fmt.Printf("[LOG] Game between %s and %s has been cleaned up.\n", game.Player1.Name, game.Player2.Name)
}

func handleSwitch(conn *net.UDPConn, addr *net.UDPAddr, id string) {
	gameKey := ""
	var player *Client
	for _, bat := range battles {
		if bat.Addr.String() != addr.String() {
			player = battles[bat.Addr.String()]
			gameKey = fmt.Sprintf("%s:%s", player.Name, battles[addr.String()].Name)
		}
	}

	game, inGame := games[gameKey]
	if !inGame {
		sendMessageToClient("No game in progress! Use @play to start a battle.", addr, conn)
		return
	}

	if addr.String() == game.Player1.Addr.String() {
		for i, poke := range game.Player1.battlePoke {
			if poke.Id == id {
				game.CurrentPoke1 = &game.Player1.battlePoke[i] // Cập nhật Pokémon hiện tại
				sendMessageToClient(fmt.Sprintf("You switched to %s.", game.CurrentPoke1.Name), game.Player1.Addr, conn)
				sendMessageToClient(fmt.Sprintf("Your opponent switched to %s.", game.CurrentPoke1.Name), game.Player2.Addr, conn)
				state = game.Player2.Addr // Chuyển lượt chơi
				return
			}
		}
	} else if addr.String() == game.Player2.Addr.String() {
		for i, poke := range game.Player2.battlePoke {
			if poke.Id == id {
				game.CurrentPoke2 = &game.Player2.battlePoke[i] // Cập nhật Pokémon hiện tại
				sendMessageToClient(fmt.Sprintf("You switched to %s.", game.CurrentPoke2.Name), game.Player2.Addr, conn)
				sendMessageToClient(fmt.Sprintf("Your opponent switched to %s.", game.CurrentPoke2.Name), game.Player1.Addr, conn)
				state = game.Player1.Addr // Chuyển lượt chơi
				return
			}
		}
	}

	if game.CurrentPoke1.PokeInfo.Hp == 0 || game.CurrentPoke2.PokeInfo.Hp == 0 {
		sendMessageToClient("Opponent needs to switch Pokémon before continuing.", addr, conn)
		return
	}

	sendMessageToClient("Invalid Pokémon ID. Please try again.", addr, conn)
}

func getDmgNumber(pAtk Pokedex, pRecive Pokedex) (int, int) {
	// Bản đồ hệ số kháng
	types := map[string]float32{
		"Normal":   pRecive.PokeInfo.TypeDefense.Normal,
		"Fire":     pRecive.PokeInfo.TypeDefense.Fire,
		"Water":    pRecive.PokeInfo.TypeDefense.Water,
		"Electric": pRecive.PokeInfo.TypeDefense.Electric,
		"Grass":    pRecive.PokeInfo.TypeDefense.Grass,
		"Ice":      pRecive.PokeInfo.TypeDefense.Ice,
		"Fighting": pRecive.PokeInfo.TypeDefense.Fighting,
		"Poison":   pRecive.PokeInfo.TypeDefense.Poison,
		"Ground":   pRecive.PokeInfo.TypeDefense.Ground,
		"Flying":   pRecive.PokeInfo.TypeDefense.Flying,
		"Psychic":  pRecive.PokeInfo.TypeDefense.Psychic,
		"Bug":      pRecive.PokeInfo.TypeDefense.Bug,
		"Rock":     pRecive.PokeInfo.TypeDefense.Rock,
		"Ghost":    pRecive.PokeInfo.TypeDefense.Ghost,
		"Dragon":   pRecive.PokeInfo.TypeDefense.Dragon,
		"Dark":     pRecive.PokeInfo.TypeDefense.Dark,
		"Steel":    pRecive.PokeInfo.TypeDefense.Steel,
		"Fairy":    pRecive.PokeInfo.TypeDefense.Fairy,
	}

	// Tính sát thương tấn công thường (Normal Attack)
	normal := float32(pAtk.PokeInfo.Atk) - float32(pRecive.PokeInfo.Def)*0.5
	if normal < 1 {
		normal = 1 // Sát thương tối thiểu
	}

	// Tính sát thương tấn công đặc biệt (Special Attack)
	typeDefense := float32(1.0) // Hệ số kháng mặc định
	for _, atkType := range pAtk.Types {
		if def, ok := types[atkType]; ok {
			if def > typeDefense { // Chọn hệ số kháng mạnh nhất
				typeDefense = def
			}
		}
	}

	special := (float32(pAtk.PokeInfo.SpAtk) * typeDefense) - float32(pRecive.PokeInfo.SpDef)*0.5
	if special < 1 {
		special = 1 // Sát thương tối thiểu
	}

	// Trả về sát thương bình thường và đặc biệt
	return int(normal), int(special)
}

func distributeExp(winner *Client, loser *Client) {
	totalExp := 0

	// Tính tổng kinh nghiệm của tất cả Pokémon trong đội thua
	for _, poke := range loser.battlePoke {
		totalExp += poke.Exp
	}

	// Kinh nghiệm thưởng cho mỗi Pokémon của đội thắng
	expReward := totalExp / len(winner.battlePoke)

	// Cập nhật kinh nghiệm và cấp độ cho từng Pokémon trong đội thắng
	for i := range winner.battlePoke {
		winner.battlePoke[i].Exp += expReward
		totalExpForNextLevel, _ := getLevelExp(winner.battlePoke[i].Level)
		if winner.battlePoke[i].Exp >= totalExpForNextLevel {
			winner.battlePoke[i].Level += 1
		}
	}
}

func getLevelExp(level int) (int, int) {

	totalExp := (level + 1) * (level + 1) * (level + 1)
	ExpAtLevel := totalExp - (level * level * level)
	if level == 1 {
		ExpAtLevel = totalExp
	}
	return totalExp, ExpAtLevel
}
