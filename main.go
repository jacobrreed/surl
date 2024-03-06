package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var baseUrl = "http://localhost:5000/"

type shortenBody struct {
	LongUrl string `json:"longUrl"`
}
type UrlDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	UrlCode   string             `bson:"urlCode"`
	LongUrl   string             `bson:"longUrl"`
	ShortUrl  string             `bson:"shortUrl"`
	CreatedAt time.Time          `bson:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt"`
}

var collection *mongo.Collection
var ctx = context.TODO()

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	dbPassword := os.Getenv("MONGO_PASSWORD")
	credential := options.Credential{
		Username: "root",
		Password: dbPassword,
	}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/").SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("db").Collection("urls")
	log.Print("DB connected")
}

func main() {
	r := gin.Default()
	r.GET("/:code", redirect)
	r.POST("/shorten", shorten)
	r.Run(":5000")
}

func shorten(c *gin.Context) {
	var body shortenBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, urlErr := url.ParseRequestURI(body.LongUrl)
	if urlErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": urlErr.Error()})
		return
	}

	urlCode, idErr := shortid.Generate()
	if idErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": idErr.Error()})
		return
	}

	var result bson.M
	queryErr := collection.FindOne(ctx, bson.D{{"urlCode", urlCode}}).Decode(&result)

	if queryErr != nil {
		if queryErr != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": queryErr.Error()})
			return
		}
	}

	if len(result) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Code in use: %s", urlCode)})
		return
	}

	var date = time.Now()
	var expires = date.AddDate(0, 0, 5)
	var newUrl = baseUrl + urlCode
	var docId = primitive.NewObjectID()

	newDoc := &UrlDoc{
		ID:        docId,
		UrlCode:   urlCode,
		LongUrl:   body.LongUrl,
		ShortUrl:  newUrl,
		CreatedAt: time.Now(),
		ExpiresAt: expires,
	}

	_, err := collection.InsertOne(ctx, newDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"newUrl":  newUrl,
		"expires": expires.Format("2006-01-02 15:04:05"),
		"db_id":   docId,
	})
}

func redirect(c *gin.Context) {
	code := c.Param("code")
	var result bson.M
	queryErr := collection.FindOne(ctx, bson.D{{"urlCode", code}}).Decode(&result)

	if queryErr != nil {
		if queryErr == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No URL with code: %s", code)})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": queryErr.Error()})
			return
		}
	}
	log.Print(result["longUrl"])
	var longUrl = fmt.Sprint(result["longUrl"])
	c.Redirect(http.StatusPermanentRedirect, longUrl)
}
