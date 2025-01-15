package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/cc-jose-nieto/go-blog-gator/internal/config"
	"github.com/cc-jose-nieto/go-blog-gator/internal/database"
	_ "github.com/lib/pq"
	"os"
)

var errNoCommandFound = errors.New("command not found")

type stateInstance struct {
	cfg *config.Config
	db  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	list map[string]func(*stateInstance, Command) error
}

func main() {
	c := *config.Read()

	db, err := sql.Open("postgres", c.DbUrl)

	state := &stateInstance{
		cfg: &c,
		db:  database.New(db),
	}
	//
	//fmt.Println(state.cfg)

	commands := Commands{
		list: make(map[string]func(*stateInstance, Command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)

	if len(os.Args) < 2 {
		fmt.Println("no command provided")
		os.Exit(1)
	}

	cmd := Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = commands.run(state, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func handlerLogin(s *stateInstance, cmd Command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("no arguments")
	}

	username := cmd.Args[0]

	_, err := s.db.GetUserByName(context.Background(), username)

	if err != nil {
		return err
	}

	s.cfg.SetUser(username)

	fmt.Printf("Welcome %s", username)

	return nil
}

func handlerRegister(s *stateInstance, cmd Command) error {

	createdUser, err := s.db.CreateUser(context.Background(), cmd.Args[0])

	if err != nil {
		return err
	}

	s.cfg.SetUser(createdUser.Name)

	fmt.Printf("User %s created successfully", createdUser.Name)

	return nil
}

func (c *Commands) register(name string, f func(*stateInstance, Command) error) {
	c.list[name] = f
}

func (c *Commands) run(s *stateInstance, cmd Command) error {

	handler, ok := c.list[cmd.Name]
	if !ok {
		return errNoCommandFound
	}

	return handler(s, cmd)
}
