package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/cc-jose-nieto/go-blog-gator/internal/config"
	"github.com/cc-jose-nieto/go-blog-gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"html"
	"net/http"
	"os"
)

var errNoCommandFound = errors.New("command not found")

type stateInstance struct {
	cfg         *config.Config
	db          *database.Queries
	currentUser *database.User
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	list map[string]func(*stateInstance, Command) error
}

func middlewareLoggedIn(handler func(s *stateInstance, cmd Command) error) func(*stateInstance, Command) error {
	return func(s *stateInstance, cmd Command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}

		if user.ID == uuid.Nil {
			return errors.New("user not found")
		}

		s.currentUser = &user

		return handler(s, cmd)
	}
}

func middlewareCheckArgs(handler func(s *stateInstance, cmd Command) error) func(*stateInstance, Command) error {
	return func(s *stateInstance, cmd Command) error {

		if len(cmd.Args) == 0 {
			return fmt.Errorf("no args provided")

		}

		return handler(s, cmd)
	}
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
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerGetAllFeeds)
	commands.register("follow", middlewareCheckArgs(middlewareLoggedIn(handlerFollow)))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareCheckArgs(middlewareLoggedIn(handlerUnfollow)))

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

func handlerAddFeed(s *stateInstance, cmd Command) error {
	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]

	//user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
	//if err != nil {
	//	return err
	//}

	createdFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{Name: feedName, Url: feedURL, UserID: s.currentUser.ID})
	if err != nil {
		return err
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{UserID: s.currentUser.ID, FeedID: createdFeed.ID})
	if err != nil {
		return err
	}

	fmt.Println(createdFeed)

	return nil
}

func handlerGetAllFeeds(s *stateInstance, cmd Command) error {
	var f []struct {
		Name     string
		Username string
	}
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserID.String())
		user, userErr := s.db.GetUserByID(context.Background(), feed.UserID)
		if userErr != nil {
			fmt.Println(userErr)
			continue
		}
		f = append(f, struct {
			Name     string
			Username string
		}{
			Username: user.Name,
			Name:     feed.Name,
		})
	}

	fmt.Println(f)
	return nil
}

func handlerFollow(s *stateInstance, cmd Command) error {

	url := cmd.Args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{UserID: s.currentUser.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}

	fmt.Println(feedFollow)
	fmt.Printf("You are now following %s\n", feed.Name)

	return nil
}

func handlerFollowing(s *stateInstance, cmd Command) error {
	fmt.Println("GetFeedFollowsByUserId")

	feeds, err := s.db.GetFeedFollowsByUserId(context.Background(), s.currentUser.ID)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("%s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *stateInstance, cmd Command) error {
	url := cmd.Args[0]
	ctx := context.Background()
	feed, err := s.db.GetFeedByUrl(ctx, url)
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollowsByUserIdAndFeedId(ctx, database.DeleteFeedFollowsByUserIdAndFeedIdParams{UserID: s.currentUser.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}

	return nil
}
