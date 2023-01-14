## Url-Shortner
This is **CRUD** operations on MongoDB written in Golang. You can **Create, Read, Update, and Delete** URLs from the MongoDB instance using HTTP requests using Postman API.

**How to Run?**

To start our project we will install our dependencies in our case will be the MongoDB go driver, gin package and shortId package:
```
  go get go.mongodb.org/mongo-driver
  go get github.com/gin-gonic/gin
  go get github.com/teris-io/shortid
```
Following the above next install the dependencies using :
**go get ./...**

Finally, run the app on port 5000:
go run .

**Endpoints:**
```
GET    /:code   
POST   /:shorten 
PUT    /:id    
DELETE /:id  

**Get URL**
```
This endpoint retrieves an entire site of a given URL.
Send a GET request to /:code:

Below is an example, Using Postman API GET send this request along with code generated in MongoDB
```
GET - http://localhost:5000/daig
```
Response:

You will get the entire site in multiple views in the response console of Postman API

**Create URL**

This endpoint inserts a URL in the url collection of the Project database.

Send a POST request to /shorten:

Below is an example, Using Postman API POST this request along with adding an URL in the Body by changing the file type(raw) into JSON
```
POST - http://localhost:5000/shorten
```
Body:
```
{
  "long_url":"https://en.wikipedia.org/wiki/Kannada"
}
```
Response:
```
{
    "db_id": "63c2dc223f6a7eb951cc8472",
    "newUrl": "http://localhost:5000/daig"
}
```
**Update URL:**

The below endpoint helps to update the existing URL in the database.

Send PUT request to /:id
```
PUT - http://localhost:5000/a753ee84-282d-4f9a-a0b6-1698121d990b
```

Response:
```
{
    "error": false,
    "message": "Long Url Updated"
}
```

**Delete URL:**

The below endpoint performs a delete URL operation in the database

Send DELETE request to /:id
```
DELETE - http://localhost:5000/a753ee84-282d-4f9a-a0b6-1698121d990b
```
Response:
```
{
    "error": false,
    "message": "Long Url Deleted"
}
```

The above procedure helps you to perform **URL CRUD operations in MongoDB.**

**Note:**

:code and :id that has been used above depend on my MongoDB creation. That is unique for every Developer.
