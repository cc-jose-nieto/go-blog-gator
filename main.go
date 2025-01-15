package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/cc-jose-nieto/go-blog-gator/internal/config"
	"github.com/cc-jose-nieto/go-blog-gator/internal/database"
	_ "github.com/lib/pq"
	"html"
	"net/http"
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
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerRSS)

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

func handlerReset(s *stateInstance, cmd Command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func handlerUsers(s *stateInstance, cmd Command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	var isCurrentUser string
	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			isCurrentUser = "(current)"
		}
		fmt.Printf("* %s %s\n", user.Name, isCurrentUser)
	}
	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	client := http.DefaultClient

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-agent", "gator")

	body, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer body.Body.Close()

	feed := &RSSFeed{}
	err = xml.NewDecoder(body.Body).Decode(feed)
	if err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	return feed, nil
}

func handlerRSS(s *stateInstance, cmd Command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	for _, item := range feed.Channel.Item {
		fmt.Printf("%s\n", item)
		//fmt.Printf("%s\n", item.Link)
		//fmt.Printf("%s\n", item.Description)
		//fmt.Printf("%s\n", item.PubDate)
	}
	return nil
}
