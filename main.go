package main

import (
	"context"
	"log"
	// "time"
	// "strconv"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017" 

type Employee struct {
	ID     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string             `json:"name"`
	Salary float64            `json:"salary"`
	Age    int                `json:"age"`
}

func Connect() error {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	db := client.Database(dbName)

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
	return nil
}


func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()

	app.Get("/employee", func(c *fiber.Ctx) error {
		query := bson.D{{}}

		cursor, err := mg.Db.Collection("employees").Find(context.Background(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		var employees []Employee
		if err := cursor.All(context.Background(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(employees)
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")
		employee := new(Employee)
		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		_, err := collection.InsertOne(context.Background(), employee)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.Status(201).JSON(employee)
	})

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		employeeID, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			return c.Status(400).SendString("Invalid employee ID")
		}

		employee := new(Employee)
		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		filter := bson.M{"_id": employeeID}
		update := bson.M{
			"$set": bson.M{
				"name":   employee.Name,
				"age":    employee.Age,
				"salary": employee.Salary,
			},
		}
		// filter := bson.D{{"_id", employeeID}}
		// update := bson.D{
		// 	{"$set", bson.D{
		// 		{"name", employee.Name},
		// 		{"age", employee.Age},
		// 		{"salary", employee.Salary},
		// 	}},
		// }

		_, err = mg.Db.Collection("employees").UpdateOne(context.Background(), filter, update)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		employee.ID = employeeID
		return c.Status(200).JSON(employee)
	})

	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		employeeID, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			return c.Status(400).SendString("Invalid employee ID")
		}

		filter := bson.D{{"_id", employeeID}}
		result, err := mg.Db.Collection("employees").DeleteOne(context.Background(), filter)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		if result.DeletedCount < 1 {
			return c.Status(404).SendString("No record deleted")
		}

		return c.Status(200).SendString("Record deleted")
	})

	log.Fatal(app.Listen(":3000"))
}
