package test

import (
	"encoding/json"
	"fmt"
)

type Zoo struct {
	Name    string
	Animals []Animal
}
type Animal struct {
	Species string
	Says    string
}

type connectionResponse struct {
	NewPlayer   bool        `json:"newPlayer"`
	GameOngoing bool        `json:"gameOngoing"`
	DisplayName string      `json:"displayName"`
	RoomId      string      `json:"roomId"`
	Data        interface{} `json:"data"`
}

func main() {
	//zoo := Zoo{"Magical Mystery Zoo",
	//	[]Animal{
	//		{"Cow", "Moo"},
	//		{"Cat", "Meow"},
	//		{"Fox", "???"},
	//	},
	//}

	response := connectionResponse{
		NewPlayer:   true,
		GameOngoing: false,
		DisplayName: "",
		RoomId:      "",
		Data:        nil,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response)
	fmt.Println(responseJson)
}
