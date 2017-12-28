package server

import (
	"log"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func waitForever() {
	ch := make(chan bool)
	<-ch
}
