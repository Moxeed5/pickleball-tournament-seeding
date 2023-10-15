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
var ctx = context.TODO()

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("Tournament").Collection("teams")
}

func containsOnlyLetters(str string) bool {
	reg := regexp.MustCompile("^[A-Za-z]+$")
	return reg.MatchString(str)
}

type Team struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	PlayerOne string             `bson:"player_1"`
	PlayerTwo string             `bson:"player_2"`
	Win       bool               `bson:"win"`
	PointTotal int               `bson:"point_total"`
	SeedNumber int				 `bson: "seed_number"`
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
						Win:       false,
						PointTotal: 0,
					}

					fmt.Printf("Successfully added %s and %s to team ID %s", playerOne, playerTwo, team.ID.Hex())

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
						fmt.Printf("Team ID: %s, Player One: %s, Player Two: %s\n", team.ID.Hex(), team.PlayerOne, team.PlayerTwo)
					}
					
					// Check if the cursor encountered any errors while iterating.
					if err := cursor.Err(); err != nil {
						log.Fatal(err) // Handle the cursor error
					}
			
					return nil
				},
			},
			
			
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createTeam(team *Team) error {
	_, err := collection.InsertOne(ctx, team)
	return err
}
