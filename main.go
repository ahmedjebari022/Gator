package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	
	"strings"
	"time"

	"github.com/ahmedjebari022/gator/internal/config"
	"github.com/ahmedjebari022/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main(){
	cfg, err := config.Read()
	if err != nil{
		fmt.Printf("%s\n",err.Error())
	}
	db, err := sql.Open("postgres",cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	s := state{
		db: dbQueries,
		cfg:&cfg,
	}
	cmdMap := make(map[string]func(*state,command)error)
	cmds := commands{
		cmdMap: cmdMap,
	}
	cmds.register("login",handlerLogin)
	cmds.register("register",handlerRegister)
	cmds.register("reset",handlerReset)
	cmds.register("users",handlerUsers)
	cmds.register("agg",handlerAgg)
	cmds.register("addfeed",middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds",handlerFeeds)
	cmds.register("follow",middlewareLoggedIn(handlerFollow))
	cmds.register("following",middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow",middlewareLoggedIn(handlerUnfollow))
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Expecting two arguments")
	}
	cmdName := args[1]
	cmdArgs := []string{}
	if len(args)> 2{
		cmdArgs = args[2:]
	}
	cmd := command{
		name :cmdName,
		arguments: cmdArgs,
	}
	err = cmds.run(&s,cmd)
	if err != nil {
		log.Fatal(err)
	}


}

type state struct{
	db *database.Queries
	cfg *config.Config
}

type command struct{
	name		string
	arguments	[]string
}

func handlerLogin(s *state,cmd command)error{
	
	if len(cmd.arguments) == 0{
		return fmt.Errorf("expecting argument: username")
	}
	_, err := s.db.GetUser(context.Background(),cmd.arguments[0])
	if err != nil{
		os.Exit(1)
	}

	s.cfg.SetUser(cmd.arguments[0])
	fmt.Printf("user has been set\n")
	return nil
}

type commands struct{
	cmdMap	map[string]	func(*state,command)error
}

func (c *commands) run(s *state, cmd command) error{
	handler, ok:= c.cmdMap[cmd.name]
	if !ok{
		return fmt.Errorf("command not in commands")
	}
	return handler(s,cmd)
}

func (c *commands) register(name string, f func(*state,command)error){
	c.cmdMap[name] = f
	
}


func handlerRegister(s *state,cmd command)error{
	if len(cmd.arguments) == 0{
		return fmt.Errorf("expecting 1 argument")
	}
	userParams := database.CreateUserParams{
		ID:uuid.New(),
		Name: cmd.arguments[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),	
	}
	_, err:= s.db.CreateUser(context.Background(),userParams)
	if err != nil{
		errMsg := err.Error()
		if strings.Contains(errMsg,"duplicate key value violates unique constraint"){
			os.Exit(1)
		}
	}
	s.cfg.SetUser(cmd.arguments[0])
	fmt.Printf("User registered succesfully\n")
	return nil
}


func handlerReset(s *state,cmd command)error{
	err := s.db.DeleteUser(context.Background())
	if err != nil{
		return err
	}
	return nil
}

func handlerUsers(s *state,cmd command)error{
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _,u := range users{
		fmt.Printf("* %s",u.Name)
		if u.Name == s.cfg.CurrentUser{
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil

}

func handlerAgg(s *state,cmd command)error{
	rssfeed, err := FetchFeed(context.Background(),"https://www.wagslane.dev/index.xml")
	if err != nil{
		return err
	}
	fmt.Printf("%s\n%s\n",rssfeed.Channel.Title,rssfeed.Channel.Description)
	for _,v := range rssfeed.Channel.Item{
		fmt.Printf("%s\n%s\n",v.Title,v.Description)
	}
	return nil
}

func handlerAddFeed(s *state,cmd command,user database.User)error{
	if len(cmd.arguments) < 2{
		os.Exit(1)
	}
	
	feed, err := s.db.CreateFeed(context.Background(),database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.arguments[0],
		Url: cmd.arguments[1],
		UserID: user.ID,
	})
	if err != nil {
		return err
	}
	ff, err := s.db.CreateFeedFollow(context.Background(),database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}



	fmt.Printf("name: %s\nurl: %s\nuser_id: %s\n",ff.FeedName,feed.Url,feed.UserID)
	return nil
}

func handlerFeeds(s *state, cmd command)error{
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _,v := range feeds{
		fmt.Printf("%s\n%s\n",v.Name,v.Name_2)
	}
	return nil
}

func handlerFollow(s *state,cmd command,user database.User)error{
	if len(cmd.arguments) > 1{
		os.Exit(1)
	}
	user, err := s.db.GetUser(context.Background(),s.cfg.CurrentUser)
	if err != nil {
		return err
	}
	feed_id, err := s.db.GetFeed(context.Background(),cmd.arguments[0])
	if err != nil {
		return err
	}
	ff, err := s.db.CreateFeedFollow(context.Background(),database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID: feed_id,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s\n%s\n",ff.FeedName,ff.UsersName)
	return nil
}

func middlewareLoggedIn(handler func(s *state,cmd command, user database.User)error) func(*state,command)error{
			return func(s *state,cmd command)error{
				u, err := s.db.GetUser(context.Background(),s.cfg.CurrentUser)
				if err != nil {
					return fmt.Errorf("this action require to log in")
				}
				return handler(s,cmd,u)
			}
}



func handlerFollowing(s *state,cmd command,user database.User)error{
		feedName, err := s.db.GetFeedFollowsForUser(context.Background(),user.ID)
		if err != nil{
			return err
		}
		fmt.Printf("%s\n",user.Name)
		for _,v := range feedName{
			fmt.Printf("%s\n",v)
		}
		return nil
}

func handlerUnfollow(s* state,cmd command,user database.User)error{
	if len(cmd.arguments) < 1 {
		os.Exit(1)
	}
	feed_id, err := s.db.GetFeed(context.Background(),cmd.arguments[0])
	if err != nil{
		return err
	}

	err = s.db.DeleteFeedFollows(context.Background(),database.DeleteFeedFollowsParams{
		UserID: user.ID,
		FeedID: feed_id,
	})
	if err != nil{
		return err
	}
	return nil

}