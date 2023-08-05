package main

import (
	"log"
	"time"

	"go.uber.org/zap"
)

func main() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("unable to create logger: %s", err)
	}

	for {
		l.Info("worker")
		time.Sleep(time.Second * 5)
	}
}
