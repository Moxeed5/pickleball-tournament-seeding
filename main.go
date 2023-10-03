package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
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

type Team struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	PlayerOne string             `bson:"player_1"`
	PlayerTwo string             `bson:"player_2"`
	Win       bool               `bson:"win"`
	PointTotal int               `bson:"point_total"`
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

					return createTeam(team)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list all the teams",
				Action: func(c *cli.Context) error {
					// Implement functionality to retrieve and display all teams
					// from the MongoDB collection. 
					// E.g., run a find() query and format the output for CLI.
					return nil
				},
			},
			// Add additional commands as per your requirement
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
