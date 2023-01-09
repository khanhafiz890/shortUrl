package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()
var baseUrl = "http://localhost:5000/"

type shortenBody struct {
	LongUrl string `json:"long_url"`
}

type UrlDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	UrlCode   string             `bson:"urlCode"`
	LongUrl   string             `bson:"longUrl"`
	ShortUrl  string             `bson:"shortUrl"`
	CreatedAt time.Time          `bson:"createdAt"`
}

type ResponseBody struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

func init() {
	clientOptons := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptons)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("test").Collection("urls")
	log.Print("DB connected")
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"message": "shorten your url"})
	})
	r.GET("/:code", redirect)
	r.POST("/shorten", shorten)
	r.PUT("/:id", updateOneUrl)
	r.DELETE("/:id", deleteOneUrl)

	r.Run(":5000")
}
func shorten(c *gin.Context) {
	var body shortenBody

	// handle error if long url not provided
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//if url is not the url itself we getting parsing error
	_, urlErr := url.ParseRequestURI(body.LongUrl)
	if urlErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": urlErr.Error()})
		return
	}

	urlCode, idErr := shortid.Generate()

	//if we getting error while generating error from Generate Method
	if idErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": idErr.Error()})
		return
	}

	var result bson.M

	queryErr := collection.FindOne(ctx, bson.D{{Key: "longUrl", Value: body.LongUrl}}).Decode(&result)

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

	var newUrl = baseUrl + urlCode
	var docId = primitive.NewObjectID()

	newDoc := &UrlDoc{
		ID:        docId,
		UrlCode:   urlCode,
		LongUrl:   body.LongUrl,
		ShortUrl:  newUrl,
		CreatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx, newDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"newUrl": newUrl,

		"db_id": docId,
	})
}

func redirect(c *gin.Context) {
	code := c.Param("code")
	var result bson.M
	queryErr := collection.FindOne(ctx, bson.D{{Key: "urlCode", Value: code}}).Decode(&result)

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
func updateOneUrl(ctx *gin.Context) {
	// id, _:=primitive.ObjectIDFromHex(urlId)
	// filter := bson.M{"_id" : id}
	// update := bson.M{"$set" : bson.M{"ShortUrl":}}
	var body shortenBody

	// handle error if long url not provided
	if err := ctx.BindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Long Url", body.LongUrl)

	//if url is not the url itself we getting parsing error
	_, urlErr := url.ParseRequestURI(body.LongUrl)
	if urlErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": urlErr.Error()})
		return
	}

	id := ctx.Param("id")
	filter := bson.M{"_id": id}
	upadte := bson.M{"$set": bson.M{"longUrl": body.LongUrl}}
	_, err := collection.UpdateOne(ctx, filter, upadte)

	if err != nil {
		fmt.Println("Error while updating..", err)
		return
	}

	//  var responsebody  = *ResponseBody
	//  responsebody.Error = false
	//  responsebody.Message = "Long Url Updated"

	ctx.JSON(http.StatusOK, gin.H{"message": "Long Url Updated", "error": false})

}
func deleteOneUrl(ctx *gin.Context) {
	id:= ctx.Param("id")
	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		fmt.Println("Error while deleting..", err)
		return
	}


	ctx.JSON(http.StatusOK, gin.H{"message": "Long Url Deleted", "error": false})

}
