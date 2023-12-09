package server

import (
	"encoding/json"
	"fmt"
	"github.com/chrisgabs/tictactoe2-backend/internal/game"
	"github.com/chrisgabs/tictactoe2-backend/internal/util"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

var g = game.GameInstance{
	Rooms:   map[string]*game.Room{},
	Players: map[string]*game.Player{},
}

//var frontendOrigin = fmt.Sprintf("%v%v:%v", FrontendProtocol, FrontendAddress, FrontendPort)

func Run() {

	log.Println("Starting Server...")

	setupRoutes()

	log.Printf("Listening to %v:%v", ServerAddress, ServerPort)
	err := http.ListenAndServeTLS(ServerAddress+":"+ServerPort, CertfilePath, KeyfilePath, nil)

	if err != nil {
		fmt.Printf("ERROR failed to listen and server %v\n", err)
		return
	}
}

func setupRoutes() {
	http.HandleFunc("/", sampleEndpoint)
	http.HandleFunc("/join", joinRoom)
	http.HandleFunc("/getCookie", provideSessionCookie)
	http.HandleFunc("/connect", connect)
	http.HandleFunc("/ws/newPlayer", createNewPlayer)                 // ws
	http.HandleFunc("/ws/existingPlayer", reinitializeExistingPlayer) // ws
	http.HandleFunc("/leave", reassignRoom)
	http.HandleFunc("/test", test)
}

func test(w http.ResponseWriter, r *http.Request) {
	if HandlePreFlight(w, r) {
		log.Println("handled preflight test")
		return
	}
	//w.Header().Set("Access-Control-Allow-Methods", "GET")
	//w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Content-Type", "application/json")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow credentials (e.g., cookies)

	//newCookie := http.Cookie{
	//	Name:  "sessionId",
	//	Value: "p_" + util.RandomString(3),
	//	Path:  "/",
	//}
	//http.SetCookie(w, &newCookie)
	//log.Printf("Cookie provided to new player: %v\n", newCookie)
	// initialize new player
	response := connectionResponse{
		NewPlayer:   true,
		GameOngoing: false,
		DisplayName: "",
		RoomId:      "",
		Data:        nil,
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Big error")
		return
	}
	//if marshaled, err := json.Marshal(response); err == nil {
	//	_, err := w.Write(marshaled)
	//	if err != nil {
	//		log.Printf("Error sending new connection response: %v\n", err)
	//		return
	//	}
	//}
}

// Deprecate
func provideSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionId")
	// cookie already exists and is valid
	if err == nil {
		// check if player already in list of players
		if _, exists := g.Players[cookie.Value]; exists {
			return
		}
	}
	newCookie := http.Cookie{
		Name:  "sessionId",
		Value: "p_" + util.RandomString(3),
		Path:  "/",
	}
	http.SetCookie(w, &newCookie)
}

func sampleEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Println("=========")
	playerID, err := retrieveAndValidateSessionCookie(r, false)
	if err != nil {
		log.Printf("Player session token: %v\n", playerID)
	} else {
		log.Printf("Sender: %v %v\n", g.Players[playerID], playerID)
	}
	log.Println("= Rooms")
	if len(g.Rooms) != 0 {
		for key, value := range g.Rooms {
			log.Printf("	[%v] %v", key, *value)
		}
	} else {
		log.Println("List of rooms are empty")
	}
	log.Println("= Players")
	if len(g.Players) != 0 {
		for key, value := range g.Players {
			log.Printf("	[%v] %v", key, *value)
		}
	} else {
		log.Println("List of players are empty")
	}
	log.Println("=========")
}

type joinRequest struct {
	RoomId string
}

type joinResponse struct {
	OpponentDisplayName string `json:"opponentDisplayName"`
	PlayerNumber        int    `json:"playerNumber"`
	RoomId              string `json:"roomId"`
}

func joinRoom(w http.ResponseWriter, r *http.Request) {
	if HandlePreFlight(w, r) {
		log.Println("handled preflight joinRoom")
		return
	}
	log.Println(" -- /joinRoom --")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow credentials (e.g., cookies)
	w.Header().Set("Content-Type", "application/json")
	playerId, err := retrieveAndValidateSessionCookie(r, false)
	if err != nil {
		log.Printf("Unable to join room: %v\n", err)
		return
	}
	if player, exists := g.Players[playerId]; exists {
		log.Printf("joining room: %v\n", player.DisplayName)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var requestData joinRequest
		if err := json.Unmarshal(body, &requestData); err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		if room, exists := g.Rooms[requestData.RoomId]; exists {
			playerNum, addedSuccessfully := g.AddPlayerToRoom(player, room)
			if addedSuccessfully {
				opponentDisplayName := room.Player1.DisplayName
				if playerNum == 1 {
					opponentDisplayName = room.Player2.DisplayName
				}
				response := joinResponse{
					OpponentDisplayName: opponentDisplayName,
					PlayerNumber:        playerNum,
				}
				marshaled, err := json.Marshal(response)
				if err != nil {
					log.Printf("ERROR marshaling join request %v\n", err)
				} else {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(marshaled)
					if err != nil {
						log.Printf("ERROR responding to join request %v\n", err)
						return
					}
					log.Printf("Player %v successfully added to room %v\n", player.DisplayName, r.RemoteAddr)
				}
			} else {
				log.Printf("ERROR: Player %v could not be added to room %v\n", player.DisplayName, room.RoomId)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("404 - Room not found"))
			if err != nil {
				return
			}
			log.Printf("ERROR: Player %v could not be added to room %v room does not exist\n", player.DisplayName, requestData.RoomId)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("404 - Player not found"))
		if err != nil {
			return
		}
		log.Printf("ERROR player attempting to join %v does not exist\n", playerId)
	}
}

type connectionResponse struct {
	NewPlayer   bool            `json:"newPlayer"`
	GameOngoing bool            `json:"gameOngoing"`
	DisplayName string          `json:"displayName"`
	RoomId      string          `json:"roomId"`
	Data        json.RawMessage `json:"data"`
}

// MarshalJSON implements the json.Marshaler interface for CustomType.
//func (c connectionResponse) MarshalJSON() ([]byte, error) {
//	// You can customize the JSON representation of your type here.
//	return json.Marshal(struct {
//		NewPlayer   bool        `json:"newPlayer"`
//		GameOngoing bool        `json:"gameOngoing"`
//		DisplayName string      `json:"displayName"`
//		RoomId      string      `json:"roomId"`
//		Data        interface{} `json:"data"`
//	}{
//		NewPlayer:   c.NewPlayer,
//		GameOngoing: c.GameOngoing,
//		DisplayName: c.DisplayName,
//		RoomId:      c.RoomId,
//		Data:        c.Data,
//	})
//}

// initial endpoint
// returns whether client's cookies is valid.

// connect Checks whether the user is already a valid user.
// if valid, return with game data
// else, provide a cookie.
// The client will have to then call /newPlayer (websocket)
func connect(w http.ResponseWriter, r *http.Request) {
	if HandlePreFlight(w, r) {
		log.Println("handled preflight connect")
		return
	}
	log.Println("-- /connect --")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow credentials (e.g., cookies)
	w.Header().Set("Content-Type", "application/json")

	isNewPlayer := false
	cookie, err := r.Cookie("sessionId")
	// cookie does not exist
	if err != nil {
		isNewPlayer = true
	} else {
		// sessionId not in list of session Ids
		_, exists := g.Players[cookie.Value]
		if !exists {
			isNewPlayer = true
		}
	}
	log.Printf("New connection with sessionId: %v isNewPlayer:%v\n", cookie, isNewPlayer)
	if isNewPlayer {
		newCookie := http.Cookie{
			Name:     "sessionId",
			Value:    "p_" + util.RandomString(3),
			Path:     "/",
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
			HttpOnly: true,
		}
		http.SetCookie(w, &newCookie)
		log.Printf("Cookie provided to new player: %v\n", newCookie)
		// initialize new player
		//w.WriteHeader(http.StatusCreated)
		response := connectionResponse{
			NewPlayer:   true,
			GameOngoing: false,
			DisplayName: "",
			RoomId:      "",
			Data:        nil,
		}
		//err1 := json.NewEncoder(w).Encode(response)
		//if err1 != nil {
		//	log.Println("ERORORO")
		//	return
		//}
		if marshaled, err := json.Marshal(response); err == nil {
			_, err := w.Write(marshaled)
			if err != nil {
				log.Printf("Error sending new connection response: %v\n", err)
				return
			}
		}
	} else { // existing player
		player, _ := g.Players[cookie.Value]
		log.Printf("Player already exists: %v\n", player.DisplayName)
		response := connectionResponse{
			NewPlayer:   false,
			GameOngoing: player.Room.GameOngoing,
			DisplayName: player.DisplayName,
			RoomId:      player.Room.RoomId,
			Data:        nil, // board data is sent on connect of websocket
		}
		if marshaled, err := json.Marshal(response); err == nil {
			_, err := w.Write(marshaled)
			if err != nil {
				log.Printf("Error sending new connection response: %v\n", err)
				return
			}
		}
		return
	}

}

//decoder := json.NewDecoder(r.Body)
//var newPlayerData newPlayerRequest
//err := decoder.Decode(&newPlayerData)
//if err != nil {
//panic(err)
//}

// A new room is also created upon creation of new player
// parse body
func createNewPlayer(w http.ResponseWriter, r *http.Request) {
	log.Println("-- /[ws] create new player --")
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		fmt.Printf("Error retrieving cookie while creating player: %v\n", err)
	}
	displayName := "new player" + cookie.Value
	log.Printf("Creating new player: %v | %v\n", cookie.Value, displayName)
	//upgrade socket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR could not upgrade socket %v\n", err)
	}
	// create new player and room
	player, created := g.CreatePlayer(cookie.Value, displayName, ws)
	if created {
		log.Printf("New Player Created: %v\n", displayName)
		room := g.CreateRoom()
		log.Printf("New Room Created: %v\n", room.RoomId)
		_, playerAddedSuccess := g.AddPlayerToRoom(player, room)
		if playerAddedSuccess {
			log.Printf("Player %v successfully added to room %v\n", displayName, room.RoomId)
		} else {
			log.Printf("ERROR: Player %v could not be added to room %v\n", displayName, room.RoomId)
		}
		go player.StartListeningToClient()
		go player.StartListeningToRoom()
	} else {
		log.Printf("ERROR Unable to create player: %v\n", displayName)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reinitializeExistingPlayer(w http.ResponseWriter, r *http.Request) {
	log.Println("-- /[ws] reinitialize --")
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		fmt.Printf("Error retrieving cookie while reinitializing player: %v\n", err)
	}
	// get player
	player := g.Players[cookie.Value]
	// upgrade socket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR could not upgrade socket %v\n", err)
	}
	player.Conn = ws
	go player.StartListeningToClient()
}

func reassignRoom(w http.ResponseWriter, r *http.Request) {
	if HandlePreFlight(w, r) {
		log.Println("handled preflight connect")
		return
	}
	log.Println("-- /reassign --")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow credentials (e.g., cookies)
	w.Header().Set("Content-Type", "application/json")
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		fmt.Printf("Error retrieving cookie while reassigning player: %v\n", err)
	}
	// get player
	p := g.Players[cookie.Value]
	room := g.CreateRoom()
	log.Printf("New Room Created: %v\n", room.RoomId)
	_, playerAddedSuccessful := g.AddPlayerToRoom(p, room)
	if playerAddedSuccessful {
		log.Printf("Player %v successfully added to room %v\n", p.DisplayName, room.RoomId)
	} else {
		log.Printf("ERROR: Player %v could not be added to room %v\n", p.DisplayName, room.RoomId)
		return
	}
	response := joinResponse{
		OpponentDisplayName: "",
		PlayerNumber:        p.PlayerNumber,
		RoomId:              room.RoomId,
	}
	marshaled, err := json.Marshal(response)
	if err != nil {
		log.Printf("ERROR marshaling join request %v\n", err)
	} else {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(marshaled)
		if err != nil {
			log.Printf("ERROR responding to join request %v\n", err)
			return
		}
		log.Printf("Player %v successfully added to room %v\n", p.DisplayName, r.RemoteAddr)
	}
}

type ErrInvalidSessionCookie string

func (e ErrInvalidSessionCookie) Error() string {
	return fmt.Sprintf("Invalid Session Cookie: %v", string(e))
}

// Validate user's session cookie
// Will only return true if:
//   - player is already in list of players
//   - creating new player

// Will return false if:
//   - cookie does not exist in request
//   - cookie is malformed
func retrieveAndValidateSessionCookie(request *http.Request, creatingNewPlayer bool) (string, error) {
	cookie, err := request.Cookie("sessionId")
	// sessionId cookie retrieved
	if err != nil {
		// sessionId does not exist
		log.Println("SessionId does not exist")
		return "", err
	} else {
		// cookie format is invalid
		if cookie.Value[0:2] != "p_" {
			log.Println("Player malformed cookie")
			return "", ErrInvalidSessionCookie(cookie.Value)
		}
		if creatingNewPlayer {
			log.Println("Creating new player")
			return cookie.Value, nil
		}
		// Check if cookie exists, delete if it does not
		if _, exists := g.Players[cookie.Value]; exists {
			log.Println("Player exists in list of players")
			return cookie.Value, nil
		} else {
			log.Println("Player does not exist in list of players")
			return "", ErrInvalidSessionCookie(cookie.Value)
		}
	}
}
