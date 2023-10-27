package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
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
	PointsLost int               `bson:"points_lost"`
	PointsWon int				`bson:"points_won"`
	SeedNumber int				 `bson:"seed_number"`
}

type ByWins []Team

func (a ByWins) Len() int {return len(a)}
	func (a ByWins) Less(i, j int) bool {return a[i].Wins < a[j].Wins}
	func (a ByWins) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPointsLost []Team

func (a ByPointsLost) Len() int {return len(a)}
func (a ByPointsLost) Less(i, j int) bool {return a[i].PointsLost < a[j].PointsLost}
func (a ByPointsLost) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPointsWon []Team

func (a ByPointsWon) Len() int {return len(a)}
func (a ByPointsWon) Less(i, j int) bool {return a[i].PointsWon > a[j].PointsWon}
func (a ByPointsWon) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }



type Match struct {
	MatchID int					`bson:"match_id"`
	TeamOne	int	`bson:"team_one"`
	TeamTwo	int	`bson:"team_two"`
	Winner	int	`bson:"winner"`
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

					teamID, err := getNextTeamID()
					if err != nil {
						return err
					}

					if playerOne == "" || playerTwo == "" {
						return errors.New("Specify names for both players")
					} else if playerOne == playerTwo{
						return errors.New("Each player's name must be unique")
					} else if containsOnlyLetters(playerOne) == false || containsOnlyLetters(playerTwo) == false {
						return errors.New("Each players name cannot contain numbers or special characters")
					}

					team := &Team{
						TeamID:        teamID,
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
				Name:    "show matches",
				Aliases: []string{"sm"},
				Usage:   "list all the matches",
				Action: func(c *cli.Context) error {
					// A filter for finding all documents. 
					// filter is an empty map that returns all documents in my Tournament collection
					filter := bson.M{}
					
					// Using context with timeout to avoid potential forever waiting.
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel() // cancel the context when the operation completes.
					
					// find requires two args, context and filter. 
					cursor, err := matchesCollection.Find(ctx, filter)
					if err != nil {
						log.Fatal(err) // Properly handle error, according to your application logic.
					}
					defer cursor.Close(ctx) // cursor is a pointer to my collection, close if it after use. 

					
			
					// Iterate through the cursor and decode each document. 
					for cursor.Next(ctx) {
						var match Match
						if err := cursor.Decode(&match); err != nil {
							log.Fatal(err) // Handle errors during cursor decoding
						}
						
						//print out info from team struct 
						fmt.Printf("Match ID: %d, Team One: %d, Team Two: %d, Winner: %d\n", match.MatchID, match.TeamOne, match.TeamTwo, match.Winner)
					}
					
					// Check if the cursor encountered any errors while iterating.
					if err := cursor.Err(); err != nil {
						log.Fatal(err) // Handle the cursor error
					}
			
					return nil
				},
			},

			{
				Name:    "match",
				Aliases: []string{"m"},
				Usage:   "Create a match between two teams",
				Action: func(c *cli.Context) error {
					teamOneId, err := strconv.Atoi(c.Args().Get(0))
					if err != nil {
						return errors.New("Invalid Team One ID")
					}
			
					teamTwoId, err := strconv.Atoi(c.Args().Get(1))
					if err != nil {
						return errors.New("Invalid Team Two ID")
					}
			
					match := &Match{
						TeamOne:    teamOneId,
						TeamTwo:    teamTwoId,
						PointsLost: 0,
						PointsWon:  0,
					}
			
					fmt.Printf("Creating match between Team %d and Team %d...\n", teamOneId, teamTwoId)
					return createMatch(match)
				},
			},
			{
				Name: "Seed",
				Aliases: []string{"s"},
				Usage: "Seed the teams based on match results",
				Action: func(cliCtx *cli.Context) error {
					
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				filter := bson.M{}
				cursor, err := collection.Find(ctx, filter)
				if err != nil {
					return err
				}
				defer cursor.Close(ctx)

				var teams ByWins
				for cursor.Next(ctx) {
				var team Team
				if err := cursor.Decode(&team); err != nil {
					return err
				}
				teams = append(teams, team)
				}
					winners, losers = winOrLose(&teams)

					sort.Sort(ByPointsLost(winners))

					sort.Sort(ByPointsWon(losers))

					seededTeams := append(winners, losers...)

					

					for index, team := range seededTeams {
						team.SeedNumber = index +1 

						fmt.Printf("Team ID: %d, Team Seed: %d\n", team.TeamID, team.SeedNumber)
					
						filter := bson.M{"team_id": team.TeamID}
            			update := bson.M{"$set": bson.M{"seed_number": team.SeedNumber}}

            			// Update the team in MongoDB
            			_, err := collection.UpdateOne(ctx, filter, update)
            			if err != nil {
                			return err
            			}	
					}	
					return nil
				},
			},
			
			{
				Name: "result",
				Aliases: []string{"r"},
				Usage: "Record the results of a match",
				Action: func(c *cli.Context) error {
					//when I get cmd line args, they come in as strings. So I need to convert the user provided match ID from a string to an int
					matchID, err := strconv.Atoi(c.Args().Get(0))
					if err !=nil {
						return err
					}
					winningTeamID, err := strconv.Atoi(c.Args().Get(1))
					if err != nil {
						return err
					}
    				PointsLost, err := strconv.Atoi(c.Args().Get(2))
					if err != nil {
						return err
					}
    				PointsWon, err := strconv.Atoi(c.Args().Get(3))
					if err != nil {
						return err
					}


					//creating filter to find the correct match in MongoDb. 

					filter := bson.M{"match_id": matchID}

					update := bson.M{
						"$set": bson.M{
							"winner": winningTeamID,
							"points_lost": PointsLost,
							"points_won": PointsWon,
						},
					}

					_, err = matchesCollection.UpdateOne(ctx, filter, update)
					if err != nil {
   						return err
					}

					var match Match
					
        err = matchesCollection.FindOne(ctx, filter).Decode(&match)
        if err != nil {
            return err
        }

        // Update the winning team
        winningTeamFilter := bson.M{"team_id": match.Winner}
        winningTeamUpdate := bson.M{
            "$inc": bson.M{
                "wins":        1,
                "points_lost": PointsLost,
				"points_won": 11,
            },
        }
        _, err = collection.UpdateOne(ctx, winningTeamFilter, winningTeamUpdate)
        if err != nil {
            return err
        }

        // Identify the losing team
        losingTeamID := match.TeamOne
        if match.TeamOne == match.Winner {
            losingTeamID = match.TeamTwo
        }

        // Update the losing team
        losingTeamFilter := bson.M{"team_id": losingTeamID}
        losingTeamUpdate := bson.M{
            "$inc": bson.M{
                "losses":      1,
                "points_won": PointsWon,
				
            },
        }
        _, err = collection.UpdateOne(ctx, losingTeamFilter, losingTeamUpdate)
        if err != nil {
            return err
        }

					
					fmt.Println("Match result updated successfully")
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

func getNextMatchID() (int, error) {
    countersCollection := client.Database("Tournament").Collection("counters")
    
    filter := bson.M{"_id": "MatchID"}
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

var byWins ByWins
var teamList ByWins
var losers ByWins
var winners ByWins

func createTeam(team *Team) error {
    _, err := collection.InsertOne(ctx, team)
	if err != nil{
    	return err
	}

	byWins = append(byWins, *team)
	return nil
}

func winOrLose(teams *ByWins) (winners, losers ByWins) {
	
	// sortedTeamList := make(ByWins, len(*teams))
	// copy(sortedTeamList, *teams)

	// sort.Sort(sortedTeamList)
	
	for _, team := range *teams {
		if team.Wins == 0 {
			losers = append(losers, team)
		} else {
			winners = append(winners, team)
		}

	}

	return  winners, losers
	
}



func createMatch(match *Match) error {
	nextID, err := getNextMatchID()
	if err != nil {
		return err
	}
	match.MatchID = nextID
	_, err = matchesCollection.InsertOne(ctx, match) // Use matchesCollection here instead of collection
	return err
}




//for seeding, I need to make rounds. The first iteration of this will be for double elim. I might make a cmd to set a var to round. It will calculate number of round based
//on number of players. Pressing the command will proceed to the next round. Matches will be associated to rounds. I'm not sure of the best way to make the 
//association. 