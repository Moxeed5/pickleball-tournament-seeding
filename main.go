package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var client *mongo.Client

var ctx = context.TODO()

var matchesCollection *mongo.Collection

func init() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("Tournament").Collection("teams")

	collection = client.Database("Tournament").Collection("teams")
	matchesCollection = client.Database("Tournament").Collection("matches") // New collection for matches
}


func containsOnlyLetters(str string) bool {
	reg := regexp.MustCompile("^[A-Za-z]+$")
	return reg.MatchString(str)
}

//using counters collection
type Team struct {
	TeamID    int                `bson:"team_id"`
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	PlayerOne string             `bson:"player_1"`
	PlayerTwo string             `bson:"player_2"`
	Wins       int               `bson:"wins"`
	Losses	   int 				 `bson:"losses"`
	PointTotal int               `bson:"point_total"`
	SeedNumber int				 `bson:"seed_number"`
}

type Match struct {
	TeamOne	primitive.ObjectID	`bson:"team_one"`
	TeamTwo	primitive.ObjectID	`bson:"team_two"`
	Winner	primitive.ObjectID	`bson:"winner"`
	PointsLost int				`bson:"points_lost"`
	PointsWon int				`bson:"points_won"`
}





func main() {
	app := &cli.App{
		Name:  "Tournament Manager",
		Usage: "A simple CLI program to create on the fly tournaments",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a team",
				Action: func(c *cli.Context) error {
					playerOne := c.Args().Get(0)
					playerTwo := c.Args().Get(1)

					if playerOne == "" || playerTwo == "" {
						return errors.New("Specify names for both players")
					} else if playerOne == playerTwo{
						return errors.New("Each player's name must be unique")
					} else if containsOnlyLetters(playerOne) == false || containsOnlyLetters(playerTwo) == false {
						return errors.New("Each players name cannot contain numbers or special characters")
					}

					team := &Team{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						PlayerOne: playerOne,
						PlayerTwo: playerTwo,
						
					}

					fmt.Printf("Successfully added %s and %s to team ID %d", playerOne, playerTwo, team.TeamID)

					return createTeam(team)
				},
			},{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list all the teams",
				Action: func(c *cli.Context) error {
					// A filter for finding all documents. 
					// filter is an empty map that returns all documents in my Tournament collection
					filter := bson.M{}
					
					// Using context with timeout to avoid potential forever waiting.
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel() // cancel the context when the operation completes.
					
					// find requires two args, context and filter. 
					cursor, err := collection.Find(ctx, filter)
					if err != nil {
						log.Fatal(err) // Properly handle error, according to your application logic.
					}
					defer cursor.Close(ctx) // cursor is a pointer to my collection, close if it after use. 
			
					// Iterate through the cursor and decode each document. 
					for cursor.Next(ctx) {
						var team Team
						if err := cursor.Decode(&team); err != nil {
							log.Fatal(err) // Handle errors during cursor decoding
						}
						//print out info from team struct 
						fmt.Printf("Team ID: %d, Player One: %s, Player Two: %s, Seed: %d\n", team.TeamID, team.PlayerOne, team.PlayerTwo, team.SeedNumber)
					}
					
					// Check if the cursor encountered any errors while iterating.
					if err := cursor.Err(); err != nil {
						log.Fatal(err) // Handle the cursor error
					}
			
					return nil
				},
			}, // Global variable for matches collection
			
			// Simplified version of the "Match" command.
			{
				Name: "match",
				Aliases: []string{"m"},
				Usage: "Create a match between two teams",
				Action: func(c *cli.Context) error {
					teamOneId := c.Args().Get(0) // Getting team IDs as input (validation is needed!)
					teamTwoId := c.Args().Get(1)
			
					// Convert hexadecimal representation of object ID to actual ObjectID.
					teamOneObjectId, err := primitive.ObjectIDFromHex(teamOneId)
					if err != nil {
						log.Fatal(err)
					}
					teamTwoObjectId, err := primitive.ObjectIDFromHex(teamTwoId)
					if err != nil {
						log.Fatal(err)
					}
			
					// Creating a new match document.
					match := &Match{
						TeamOne: teamOneObjectId,
						TeamTwo: teamTwoObjectId,
						PointsLost: 0, // Initial values can be set to zero, or take input from user
						PointsWon: 0,
					}
			
					fmt.Printf("Creating match between Team %s and Team %s...\n", teamOneId, teamTwoId)
					return createMatch(match) // Separate function to handle the creation of a match
				},
			},{
				Name: "result",
				Aliases: []string{"r"},
				Usage: "Mark the wins and losses for a team",
				Action: func(c *cli.Context) error {
					teamOneId := c.Args().Get(0) // Getting team IDs as input (validation is needed!)
					teamTwoId := c.Args().Get(1)



				},

			},
			
			
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getNextTeamID() (int, error) {
    countersCollection := client.Database("Tournament").Collection("counters")
    
    filter := bson.M{"_id": "teamID"}
    update := bson.M{"$inc": bson.M{"seq": 1}}
    opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
    
    var result struct {
        Seq int `bson:"seq"`
    }
    
    err := countersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
    
    if err != nil {
        return 0, err
    }
    
    return result.Seq, nil
}

func createTeam(team *Team) error {
    nextID, err := getNextTeamID()
    if err != nil {
        return err
    }
    team.TeamID = nextID
    _, err = collection.InsertOne(ctx, team)
    return err
}


func createMatch(match *Match) error {
	_,err := collection.InsertOne(ctx, match)
	return err
}


//next create func to calculate seed number for each team based on w/l and points