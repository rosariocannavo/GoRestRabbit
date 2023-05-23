package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type persona struct {
	Nome    string `json:"nome"`
	Cognome string `json:"cognome"`
}

func write(message persona) {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://myuser:mypassword@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	// Declare a queue
	queue, err := ch.QueueDeclare(
		"my-queue", // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	// Publish a message to the queue
	bytes, err := json.Marshal(message)
	err = ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bytes, //[]bytes(message)
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %s", err)
	}

	log.Printf("Message sent to queue '%s'", queue.Name)
}

func consume() {

	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://myuser:mypassword@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	// Declare a queue
	queue, err := ch.QueueDeclare(
		"my-queue", // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("Failed to consume messages from queue: %s", err)
	}

	for msg := range msgs {
		log.Printf("Received a message: %s", msg.Body)

	}

}

func mongoConnect() *mongo.Collection {
	clientOptions := options.Client().ApplyURI("mongodb://myuser:mypassword@mongo_db:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return nil
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Println("Error pinging MongoDB:", err)
		return nil
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("testdb").Collection("people")

	return collection
}

func writeOnQueue(c *gin.Context) {
	//connecto to mongodb
	var collection *mongo.Collection = mongoConnect()

	nome := c.Param("nome")
	cognome, ok := c.GetQuery("cognome") //query param ?

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid or missing id query parameter"})
		return
	}

	var p persona = persona{Nome: nome, Cognome: cognome}
	write(p)
	c.IndentedJSON(http.StatusCreated, gin.H{"message": p})

	//write on database
	res, err := collection.InsertOne(context.Background(), p)
	if err != nil {
		fmt.Println("Error inserting document:", err)
		return
	}

	// Print the ID of the inserted document
	fmt.Println("Inserted document ID:", res.InsertedID)

}

func readFromQueue(c *gin.Context) {
	consume()
}

func printSomething(c *gin.Context) {
	var p persona = persona{Nome: "nome", Cognome: "cognome"}

	defer fmt.Print("test")
	c.IndentedJSON(http.StatusCreated, gin.H{"message": p})
}

func main() {

	// for i := 0; i < 10; i++ {
	// 	writer(strconv.Itoa(i))
	// }

	// for i := 0; i < 10; i++ {
	// 	consume()
	// }

	router := gin.Default()
	router.GET("/write:nome", writeOnQueue) //localhost:8080/write:saro?cognome=cannavo
	router.GET("/read", readFromQueue)
	router.GET("/test", printSomething)

	router.Run("0.0.0.0:8080")
}
