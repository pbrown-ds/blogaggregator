package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DuperSoup/blogaggregator/internal/config"
	"github.com/DuperSoup/blogaggregator/internal/database"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

type state struct {
	cfg   *config.Config
	db    *database.Queries
	sqlDB *sql.DB
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

// Replaces the currently logged in user with the username provided.
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("please provide a username after the login command\n")
	}

	username := cmd.args[0]
	// Check if user already in database
	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		fmt.Printf("User does not exist in database: %v\n", err)
		os.Exit(1)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("Error setting user in handlerLogin: %v\n", err)
	}

	fmt.Printf("%s has been set as the current user\n", username)

	return nil
}

// Registers a new user with the provided username.
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("please provide a username after the register command\n")
	}

	username := cmd.args[0]

	// Check if user already in database
	_, err := s.db.GetUser(context.Background(), username)
	if err == nil {
		fmt.Printf("User already exists in database: %v\n", err)
		os.Exit(1)
	}

	// Set user parameters
	userParms := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}

	// Create user with parameters
	user, err := s.db.CreateUser(context.Background(), userParms)
	if err != nil {
		return fmt.Errorf("Error creating user in handlerRegister: %v\n", err)
	}

	// Set the current user in the config to the created user
	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("Error setting user to the created user in handlerRegister: %v\n", err)
	}

	// Print info to console
	fmt.Printf("New user %s was created.\n", username)
	fmt.Println(user.ID)
	fmt.Println(user.CreatedAt)
	fmt.Println(user.UpdatedAt)
	fmt.Println(user.Name)

	return nil
}

// Resets the database to run goose down then up migrations
func handlerReset(s *state, cmd command) error {
	goose.SetDialect("postgres")
	if err := goose.DownTo(s.sqlDB, "sql/schema", 0); err != nil {
		return err
	}
	if err := goose.Up(s.sqlDB, "sql/schema"); err != nil {
		return err
	}
	fmt.Println("Database reset and migrations applied.")
	return nil
}

// Returns a list of all the users, printed in a specific format
func handlerGetUsers(s *state, cmd command) error {
	// Delete all users from table
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Unable to get all users: %v\n", err)
		os.Exit(1)
	}

	for _, user := range users {
		// User matches the currently logged in user
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

// Aggregates feeds by fetching the RSS Feeds, parsing them, and printing the posts to the console all in a long-running loop. Provide a time between requests.
func handlerAgg(s *state, cmd command) error {
	// Get the time between requests
	if len(cmd.args) == 0 {
		return fmt.Errorf("please provide a time between requests after the Agg command in this format: #units (ex. 1m for 1 minute)\n")
	}

	timebr := cmd.args[0]

	time_between_requests, err := time.ParseDuration(timebr)
	if err != nil {
		fmt.Printf("Error parsing time between requests: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Collecting feeds every %s\n", timebr)

	// Start ticker wit the time between requests
	ticker := time.NewTicker(time_between_requests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

	// // provided test url
	// url := "https://www.wagslane.dev/index.xml"
	//
	// feed, err := fetchFeed(context.Background(), url)
	// if err != nil {
	// 	fmt.Printf("Unable fetch feed at url: %v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Print(feed)

	return nil
}

// Adds a feed with the provided feed name  and url to the current user's followed feeds.
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) <= 1 {
		return fmt.Errorf("please provide a feed name and url after the AddFeed  command\n")
	}

	feed_name := cmd.args[0]
	feed_url := cmd.args[1]

	// // Get current user and their ID
	// current_username := s.cfg.CurrentUserName
	// current_user, err := s.db.GetUser(context.Background(), current_username)
	// if err != nil {
	// 	fmt.Printf("Current user not obtained in AddFeed: %v\n", err)
	// 	os.Exit(1)
	// }

	// Set feed parameters
	feedParms := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feed_name,
		Url:       feed_url,
		UserID:    user.ID,
	}

	// Create Feed
	feed, err := s.db.CreateFeed(context.Background(), feedParms)
	if err != nil {
		fmt.Printf("Unable to create feed: %v\n", err)
		os.Exit(1)
	}

	// Set feed_follow parameters
	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	// Create feed_follow record for the current user
	_, err = s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		fmt.Printf("Unable to create feed_follows record: %v\n", err)
		os.Exit(1)
	}

	// Print info of feed to console
	fmt.Print("New feed was created.\n")
	fmt.Println(feed.ID)
	fmt.Println(feed.CreatedAt)
	fmt.Println(feed.UpdatedAt)
	fmt.Println(feed.Name)
	fmt.Println(feed.Url)
	fmt.Println(feed.UserID)

	return nil
}

// Prints all the feeds in the database to the console
func handlerFeeds(s *state, cmd command) error {
	// Get the feeds from the feeds table
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Unable to get all feeds: %v\n", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		// Fetch the user that created that feed
		user, err := s.db.GetUserID(context.Background(), feed.UserID)
		if err != nil {
			fmt.Printf("Unable to get fetch user based on userID: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("* %s\n", feed.Name)
		fmt.Printf("* %s\n", feed.Url)
		fmt.Printf("* %s\n\n", user.Name)
	}

	return nil
}

// Creates a new feed follow record for the current user and prints the name of the feed and current user once the record is created.
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("please provide a url after the Follows command\n")
	}

	feed_url := cmd.args[0]

	// Get the feed from the feeds table with that url
	feed, err := s.db.GetFeedByURL(context.Background(), feed_url)
	if err != nil {
		fmt.Printf("Unable to get feed with that url: %v\n", err)
		os.Exit(1)
	}

	// // Get current user and their ID
	// current_username := s.cfg.CurrentUserName
	// current_user, err := s.db.GetUser(context.Background(), current_username)
	// if err != nil {
	// 	fmt.Printf("Current user not obtained in Follow: %v\n", err)
	// 	os.Exit(1)
	// }

	// Set feed_follow parameters
	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		fmt.Printf("CreateFeedFollows failed: %v\n", err)
		os.Exit(1)
	}

	// Print to console
	fmt.Print("New feed_follow created:\n")
	fmt.Printf("* %s\n", feed.Name)
	fmt.Printf("* %s\n\n", user.Name)

	return nil
}

// Prints all the names of the feeds the current user is following
func handlerFollowing(s *state, cmd command, user database.User) error {
	// // Get current user and their ID
	// current_username := s.cfg.CurrentUserName
	// current_user, err := s.db.GetUser(context.Background(), current_username)
	// if err != nil {
	// 	fmt.Printf("Current user not obtained in Follow: %v\n", err)
	// 	os.Exit(1)
	// }

	feed_follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Printf("Could not get feed follows for current user: %v\n", err)
		os.Exit(1)
	}

	// Print to console the feeds
	fmt.Printf("%s is following these feeds:\n", user.Name)
	for _, feed_follow := range feed_follows {
		fmt.Printf("* %s\n", feed_follow.FeedName)
	}

	return nil
}

// Takes a feed's URL as an argument and unfollows it for the current user.
func handlerUnfollow(s *state, cmd command, user database.User) error {
	// Get the user and feed arguments
	if len(cmd.args) == 0 {
		return fmt.Errorf("please provide a feed url after the Unfollow command\n")
	}

	feed_url := cmd.args[0]

	// Fetch the feed to get its ID
	feed, err := s.db.GetFeedByURL(context.Background(), feed_url)
	if err != nil {
		fmt.Printf("Feed with that url not obtained in Unfollow: %v\n", err)
		os.Exit(1)
	}

	// Set DeleteFeedFollowParams parameters
	deleteFeedFollowParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), deleteFeedFollowParams)
	if err != nil {
		fmt.Printf("Could not unfollow that feed for current user: %v\n", err)
		os.Exit(1)
	}

	// Print to console the feeds
	fmt.Printf("%s is no longer following %s.\n", user.Name, feed.Name)

	return nil
}

// Takes an optional "limit" parameter and prints posts to the terminal up to the limit
func handlerBrowse(s *state, cmd command, user database.User) error {
	// Set the limit to 2 by default
	limit := 2
	if len(cmd.args) == 1 {
		specifiedLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			fmt.Printf("Error setting browse limit: %v\n", err)
			os.Exit(1)
		}
		limit = specifiedLimit
	}

	postsForUserParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	// Print posts to terminal based on user
	posts_for_user, err := s.db.GetPostsForUser(context.Background(), postsForUserParams)
	if err != nil {
		fmt.Printf("Error getting posts for current user: %v\n", err)
		os.Exit(1)
	}

	for _, post := range posts_for_user {
		fmt.Printf("Title: %v\n", post.Title)
		fmt.Printf("URL: %v\n", post.Url)
		fmt.Printf("Description: %v\n", post.Description)
		fmt.Printf("Publish Date: %v\n\n", post.PublishedAt)
	}

	return nil
}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(*state, command) error {

	return func(s *state, cmd command) error {
		// Ensure a user is configured
		username := s.cfg.CurrentUserName
		if username == "" {
			return fmt.Errorf("no user is currently logged in")
		}

		// Fetch the user from the database
		user, err := s.db.GetUser(context.Background(), username)
		if err != nil {
			return fmt.Errorf("could not get current user: %w", err)
		}

		// Call the wrapped handler with the authenticated user
		return handler(s, cmd, user)
	}
}

func (c *commands) run(s *state, cmd command) error {
	cmdToRun, ok := c.commandMap[cmd.name]
	if !ok {
		return fmt.Errorf("no command exists with that name\n")
	}

	return cmdToRun(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) error {
	// Check if command name already in map
	_, ok := c.commandMap[name]
	if ok { // command name is in map
		return fmt.Errorf("command name is already being used\n")
	}

	// Set new command name and function in map
	c.commandMap[name] = f

	return nil
}

// Gets the next feed to fetch from the DB, marks it as fetched, then fetches the feed using the provided URL, and iterates over the items in the feed and prints their titles to the console
func scrapeFeeds(s *state) error {
	// Get next feed to fetch
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Error getting next feed to fetch: %v\n", err)
		os.Exit(1)
	}

	// Mark feed as fetched
	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("Error marking feed as fetched: %v\n", err)
		os.Exit(1)
	}

	// Fetch the feed content using its url
	rss_feed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Error fetching RSS Feed: %v\n", err)
		os.Exit(1)
	}

	// Iterate over items in the RSSFeed and save posts to database
	for _, item := range rss_feed.Channel.Item {
		// Parse publish date if needed
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			publishedAt = time.Now()
		}

		// Set feed parameters
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		// Create post with params
		_, err = s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			if strings.Contains(err.Error(), "unique constraint") {
				// post with that url already exists, skip it
				continue
			}
			// some other error, log it
			log.Printf("Error creating post: %v\n", err)
		}
	}

	return nil
}
