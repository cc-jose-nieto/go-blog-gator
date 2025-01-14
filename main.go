package main

import (
	"errors"
	"fmt"
	"github.com/cc-jose-nieto/go-blog-gator/internal/config"
	"os"
)

var errNoCommandFound = errors.New("command not found")

type stateInstance struct {
	cfg *config.Config
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

	state := &stateInstance{
		cfg: &c,
	}

	fmt.Println(state.cfg)

	commands := Commands{
		list: make(map[string]func(*stateInstance, Command) error),
	}

	commands.register("login", handlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("no command provided")
		os.Exit(1)
	}

	cmd := Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err := commands.run(state, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func handlerLogin(s *stateInstance, cmd Command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("no arguments")
	}

	s.cfg.SetUser(cmd.Args[0])

	fmt.Printf("Welcome %s", cmd.Args[0])

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
