package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	commands "github.com/finfreezer/blogAggregator/internal/commands"
	config "github.com/finfreezer/blogAggregator/internal/config"
	"github.com/finfreezer/blogAggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	newConf := config.Config{}
	newConf = config.Read()
	newState := commands.State{Config: &newConf}
	newCommands := commands.Commands{
		CommandMap: make(map[string]func(*commands.State, commands.Command) error),
	}
	newCommand := commands.Command{}
	addCommands(&newCommands)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Missing argument, exiting.")
		os.Exit(1)
	}
	if len(args) > 2 {
		newCommand.Name = args[1]
		newCommand.Args = args[2:]
	} else {
		newCommand.Name = args[1]
	}
	db, err := sql.Open("postgres", newConf.DbURL)
	if err != nil {
		os.Exit(1)
	}
	dbQueries := database.New(db)
	newState.Db = dbQueries
	err = newCommands.Run(&newState, newCommand)
	if err != nil {
		log.Fatal(err)
	}
}

func addCommands(newCommands *commands.Commands) {
	newCommands.Register("register", commands.HandlerRegister)
	newCommands.Register("login", commands.HandlerLogin)
	newCommands.Register("reset", commands.HandlerReset)
	newCommands.Register("users", commands.HandlerListUsers)
	newCommands.Register("agg", commands.HandlerFetchFeed)
	newCommands.Register("addfeed", commands.HandlerAddFeed)
	newCommands.Register("feeds", commands.HandlerListFeeds)
}
