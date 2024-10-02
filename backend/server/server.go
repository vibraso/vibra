package server

import (
	"log"

	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	log.Println("Creating new Server instance")
	server := &Server{
		Router: mux.NewRouter(),
	}
	log.Println("Created new mux Router")
	return server
}