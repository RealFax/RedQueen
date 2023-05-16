package main

import (
	"github.com/RealFax/RedQueen/config"
	"log"
	"os"
)

func main() {
	args, err := config.ReadFromArgs(os.Args...)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(args)
}
