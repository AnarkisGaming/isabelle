package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"get.cutie.cafe/isabelle/config"
	"get.cutie.cafe/isabelle/discord"
	"get.cutie.cafe/isabelle/github"
	"get.cutie.cafe/isabelle/mail"
)

func main() {
	fmt.Println("Isabelle")
	fmt.Println("Copyright (c) 2023 Anarkis Gaming/Cutie Cafe")
	fmt.Println("This program is free software; see the LICENSE file for details.")
	fmt.Println("")

	if err := config.Init(); err != nil {
		panic(fmt.Errorf("error reading config: %v", err))
	}

	github.Init()

	if err := discord.Connect(); err != nil {
		panic(fmt.Errorf("error connecting to Discord: %v", err))
	}

	mail.Schedule()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Shutting down...")
}
