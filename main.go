package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/DuperSoup/blogaggregator/internal/config"
	"github.com/DuperSoup/blogaggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize config and store in new instance of state struct
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Problem reading config: %v\n", err)
		os.Exit(1)
	}
	// st := state{
	// 	db:  nil,
	// 	cfg: &cfg,
	// }

	// Load in database URL to config struct
	dbURL := cfg.DBURL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Problem opening sql connection: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	programState := &state{
		db:    dbQueries,
		sqlDB: db,
		cfg:   &cfg,
	}

	// Initialize empty map of handler functions
	cmds := commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	// Register the commands
	err = cmds.register("login", handlerLogin)
	if err != nil {
		fmt.Printf("Problem registering login command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("register", handlerRegister)
	if err != nil {
		fmt.Printf("Problem registering register command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("reset", handlerReset)
	if err != nil {
		fmt.Printf("Problem registering reset command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("users", handlerGetUsers)
	if err != nil {
		fmt.Printf("Problem registering get_users command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("agg", handlerAgg)
	if err != nil {
		fmt.Printf("Problem registering agg command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	if err != nil {
		fmt.Printf("Problem registering add_feed command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("feeds", handlerFeeds)
	if err != nil {
		fmt.Printf("Problem registering feeds command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("follow", middlewareLoggedIn(handlerFollow))
	if err != nil {
		fmt.Printf("Problem registering follow command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("following", middlewareLoggedIn(handlerFollowing))
	if err != nil {
		fmt.Printf("Problem registering following command: %v\n", err)
		os.Exit(1)
	}
	err = cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	if err != nil {
		fmt.Printf("Problem registering unfollow command: %v\n", err)
		os.Exit(1)
	}

	// Parse arguments
	term_args := os.Args
	if len(term_args) < 2 {
		fmt.Println("no arguments provided for the command.")
		os.Exit(1)
	}

	// Create new command struct and populate with results of os.Args
	cmd := command{
		term_args[1],
		term_args[2:], // all arguments after the name
	}

	// Run the commands in the command struct
	err = cmds.run(programState, cmd)
	if err != nil {
		fmt.Printf("Problem running command: %v\n", err)
		os.Exit(1)
	}

}
