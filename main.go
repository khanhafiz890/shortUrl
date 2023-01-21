package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Creating a Mongo DB Collection
var collection *mongo.Collection
var ctx = context.TODO()

// Creating a LocalHost url
var baseUrl = "http://localhost:5000/"

// Creating a Structure
type shortenBody struct {
	LongUrl string `json:"long_url"`
}

type SubmissionBody struct {
	LanguageId string `json:"language_id"`
	SourceCode string `json:"source_code"`
	StdInput   string `json:"std_in"`
}
type UrlDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	UrlId     string             `bson:"url_id"`
	UrlCode   string             `bson:"urlCode"`
	LongUrl   string             `bson:"longUrl"`
	ShortUrl  string             `bson:"shortUrl"`
	CreatedAt time.Time          `bson:"createdAt"`
}

type TokenBody struct {
	TokenID    primitive.ObjectID `bson:"token_id"`
	Token      string             `bson:"token"`
	LanguageId int                `json:"language_id"`
	SourceCode string             `json:"source_code"`
	Stdin      string             `json:"stdin"`
}
type ResTokenBody struct {
	Token      string `bson:"token"`
	LanguageId int    `json:"language_id"`
	SourceCode string `json:"source_code"`
	Stdin      string `json:"stdin"`
}

type ResponseBody struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

type SubmissionRequest struct {
	TokenID    primitive.ObjectID `bson:"token_id"`
	LanguageId int                `json:"language_id"`
	SourceCode string             `json:"source_code"`
	Stdin      string             `json:"stdin"`
	Token      string             `json:"token"`
}

// Connecting Mongo DB to localhost
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

// main function for CRUD operations
func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"message": "shorten your url"})
	})
	r.GET("/:code", redirect)               // Read operation End Point for shorten url
	r.POST("/shorten", shorten)             // Create Operation End Point for shorten url
	r.PUT("/:id", updateOneUrl)             // Update Operation End Point for shorten url
	r.DELETE("/:id", deleteOneUrl)          // Delete Operation End Point for shorten url
	r.POST("/submission", handleSubmission) // Create operation End Point for rapid api(Judge CE 0)
	r.GET("/submission", getSubmission)     // Read operation End Point for rapid api(Judge CE 0)

	r.Run(":5000") //  http port
}

// " /submission "  Create function End Point for rapid api(Judge CE 0)
func handleSubmission(c *gin.Context) {

	// Structre of SubmissionBody
	var body SubmissionBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// judge0-ce url

	url := "https://judge0-ce.p.rapidapi.com/submissions?base64_encoded=true&wait=true&fields=*"

	// maping to get the data that passed in body
	payload := map[string]string{"language_id": (body.LanguageId), "source_code": body.SourceCode, "stdin": body.StdInput}

	// Marshalling the json data
	json_data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error while marshaling")
	}
	fmt.Println("byte formate", string(json_data))

	// sending request
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(json_data))

	// Request header fields
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-RapidAPI-Key", "b6ee9d55camshe7ac66ecbd9ba32p10b88ajsn4d6a39a75634")
	req.Header.Add("X-RapidAPI-Host", "judge0-ce.p.rapidapi.com")

	// reciving the response
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	fmt.Println(res.Body)

	decoder := json.NewDecoder(res.Body)
	fmt.Println(decoder)

	//structure for ResTokenBody
	var tokenBody ResTokenBody
	err = decoder.Decode(&tokenBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("-------------------")
	fmt.Println(tokenBody.Token)
	fmt.Println("---------------------")

	// Creating a mongo DB for token storing

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	clientSource, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = clientSource.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = clientSource.Database("test").Collection("data")
	var docId = primitive.NewObjectID()
	newDoc := &TokenBody{

		TokenID:    docId,
		Token:      tokenBody.Token,
		LanguageId: tokenBody.LanguageId,
		SourceCode: tokenBody.SourceCode,
		Stdin:      tokenBody.Stdin,
	}
	s, err := collection.InsertOne(ctx, newDoc)
	fmt.Println(s)

	log.Print("DB set for token connected")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// http method status : 201
	c.JSON(http.StatusCreated, gin.H{
		"Message":     "Token is created",
		"token":       tokenBody.Token,
		"language_id": tokenBody.LanguageId,
		"source_code": tokenBody.SourceCode,
		"stdin":       tokenBody.Stdin,
	})

}

// " /submission/:token "  Create function End Point for rapid api(Judge CE 0)
func getSubmission(c *gin.Context) {

	// Quary param for token
	token := c.Param("token")
	fmt.Println("token =>", token)
	clientOptons := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptons)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	//Creating a Collection
	collection = client.Database("test").Collection("data")
	filter := bson.M{"token": token}
	var getReq SubmissionRequest
	err = collection.FindOne(context.TODO(), filter).Decode(&getReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(getReq)

	//Status OK response: 200
	c.JSON(200, gin.H{
		"error": false,

		"message": "Data retrived successfully",
		"data":    getReq,
	})

}

// Create Operation function End Point for shorten url
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
	// it will short the url param not more than 4words
	shorturlid, idErr := shortid.Generate()
	urlCode := shorturlid[0:4]

	//if we getting error while generating error from Generate Method
	if idErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": idErr.Error()})
		return
	}
	// map[string]interface{}
	var result bson.M
	// Query to fetch data
	queryErr := collection.FindOne(ctx, bson.D{{Key: "longUrl", Value: body.LongUrl}}).Decode(&result)
	if queryErr != nil {
		if queryErr != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": queryErr.Error()})
			return
		}
	}
	fmt.Println(result)

	if len(result) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Code in use: %s", urlCode)})
		return
	}
	//concatinating urlcode string to baseUrl
	var newUrl = baseUrl + urlCode
	//creating object id
	var docId = primitive.NewObjectID()
	// Structure for UrlDoc
	newDoc := &UrlDoc{
		ID:        docId,
		UrlCode:   urlCode,
		LongUrl:   body.LongUrl,
		ShortUrl:  newUrl,
		CreatedAt: time.Now(),
		UrlId:     uuid.New().String(),
	}
	//insrt that data into Mongo DB
	_, err := collection.InsertOne(ctx, newDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// http method status : 201
	c.JSON(http.StatusCreated, gin.H{
		"newUrl": newUrl,

		"db_id": docId,
	})
}

// Update  function End Point for shorten url
func redirect(c *gin.Context) {

	//fetching param from url
	code := c.Param("code")

	// map[string]interface{}
	var result bson.M

	//fetching data from database
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

	//Checking response moved to url
	c.Redirect(http.StatusPermanentRedirect, longUrl)
}
func updateOneUrl(ctx *gin.Context) {

	//structure for shortenBody
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

	//update the new url and store it in DB
	id := ctx.Param("id")
	fmt.Println("ID=>", id)
	filter := bson.D{{Key: "url_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "longUrl", Value: body.LongUrl}}}}
	_, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		fmt.Println("Error while updating..", err)
		return
	}

	//response status 200
	ctx.JSON(http.StatusOK, gin.H{"message": "Long Url Updated", "error": false})

}

// Delete function End Point for shortening url
func deleteOneUrl(ctx *gin.Context) {

	//deleting the url using uuid
	id := ctx.Param("id")
	fmt.Println("ID=>", id)
	filter := bson.D{{Key: "url_id", Value: id}}
	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		fmt.Println("Error while deleting..", err)
		return
	}

	//status response 200
	ctx.JSON(http.StatusOK, gin.H{"message": "Long Url Deleted", "error": false})

}
