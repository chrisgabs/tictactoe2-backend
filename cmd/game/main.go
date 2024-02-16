package main

import (
	"flag"
	"github.com/chrisgabs/tictactoe2-backend/internal/server"
	"log"
	"os"
)

func main() {
	devMode := flag.Bool("dev", false, "Whether to run the application in development mode or not.")
	flag.Parse()

	serverAddress := server.ServerAddress
	serverPort := server.ServerPort
	certfilePath := server.CertfilePath
	keyfilePath := server.KeyfilePath

	if !*devMode {
		log.Println("Running prod")
		serverAddress = os.Getenv("SERVER_ADDRESS")
		serverPort = os.Getenv("SERVER_PORT")
		certfilePath = os.Getenv("CERTFILE_PATH")
		keyfilePath = os.Getenv("KEYFILE_PATH")
		log.Println("ENVIRONMENT VARIABLES:")
		log.Println(serverAddress)
		log.Println(serverPort)
		log.Println(certfilePath)
		log.Println(keyfilePath)
	}

	server.Run(serverAddress, serverPort, certfilePath, keyfilePath)
}
