package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	c "github.com/finfreezer/blogAggregator/internal/config"
	"github.com/finfreezer/blogAggregator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	Db     *database.Queries
	Config *c.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	CommandMap map[string]func(*State, Command) error
}

func HandlerListFeeds(s *State, cmd Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Something went wrong fetching feeds: %s", err.Error())
	}

	for _, feed := range feeds {
		fmt.Printf("%s, %s\n", feed.Name, feed.Url)
		user, err := s.Db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return errors.New("Problem fetching user.")
		}
		fmt.Println(user.Name)
	}
	return nil
}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return errors.New("Not enough arguments.")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]
	user, err := s.Db.GetUser(context.Background(), s.Config.CurrentUserName)

	if err != nil {
		return fmt.Errorf("Something went wrong: %s\n", err.Error())
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}
	_, err = s.Db.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Something went wrong: %s\n", err.Error())
	}
	return nil
}

func HandlerFetchFeed(s *State, cmd Command) error {
	//feedURL := cmd.Args[0]
	newFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	fmt.Println(newFeed.Channel.Item)
	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}
	return nil
}

func HandlerListUsers(s *State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Something broke: %s\n", err.Error())
	}

	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.Db.DeleteUsers(context.Background())
	if err == nil {
		fmt.Println("Deletion succesful.")
		return nil
	} else {
		fmt.Println(err)
		return fmt.Errorf("Deletion unsuccesful: Error %w\n", err)
	}
}

func HandlerRegister(s *State, cmd Command) error {
	argName := cmd.Args[0]
	if len(cmd.Args) == 0 {
		return errors.New("Missing username.")
	}
	name, err := s.Db.GetUser(context.Background(), argName)

	if err == sql.ErrNoRows {
		params := database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      argName,
		}
		s.Db.CreateUser(context.Background(), params)
		s.Config.SetUser(argName)
		return nil

	} else if name.Name == argName {
		fmt.Println("Name already exists.")
		os.Exit(1)
	}
	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	argName := cmd.Args[0]
	if len(cmd.Args) == 0 {
		return errors.New("Missing username.")
	}
	_, err := s.Db.GetUser(context.Background(), argName)
	if err != nil {
		return fmt.Errorf("No 'user: %s' found in database, login unavailable.\n", argName)
	}
	s.Config.SetUser(argName)
	fmt.Printf("User '%s' has been set.\n", argName)

	return nil
}

func (c *Commands) Run(s *State, cmd Command) error {
	if command, ok := c.CommandMap[cmd.Name]; ok {
		err := command(s, cmd)
		return err
	} else {
		return errors.New("Command not found")
	}
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.CommandMap[name] = f
}
