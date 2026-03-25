gator (short for blogaggregator) is a command-line interface tool that can aggregate RSS Feeds from blog sites and then display their contents. You will need Postgres and Go installed to run the program.

Follow these steps to install and run the gator program:
1. In your CLI, type 'go install github.com/DuperSoup/blogaggregator@latest'
2. Open gatorconfig.json and enter in your database connection details into:

{
  "db_url": "postgres://username:@localhost:5432/database?sslmode=disable"
}

    a. `username`: Your Postgres username
    b. The password slot (between : and @) — if you have one
    c. 5432 — the port, if you are not using the default
    d. database — the name of your Postgres database

Below is an example:

{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
}

3. To run the program, type into your CLI 'gator command arguements'. There should be a space between each word. Replace 'command' with your desired command and 'arguements' with the arguements if there are any.

Below is a list of the available commands:

    1. login: Sets the current user with the username provided. Arguments: username.
    2. register: Registers a new user with the provided username. Arguments: username.
    3. reset: Resets the database to run goose down then up migrations
    4. users: Returns a list of all the users, printed in a specific format
    5. agg: Aggregates feeds by fetching the RSS Feeds, parsing them, and printing the posts to the console all in a long-running loop. Provide a time between requests. Arguments: time between requets in this format: x# (ex. 1m for 1 minute).
    6. addfeed: Adds a feed with the provided feed name to the current user's followed feeds. Arguments: feed name, feed url.
    7. feeds: Prints all the feeds in the database to the console.
    8. follow: Creates a new feed follow record for the current user and prints the name of the feed and current user once the record is created. Arguments: feed url.
    9. following: Prints all the names of the feeds the current user is following.
    10. unfollow: Takes a feed's URL as an argument and unfollows it for the current user. Arguments: feed url.
    11. browse: Takes an optional "limit" parameter and prints posts to the terminal up to the limit. Arguments: limit (an integer like 3).