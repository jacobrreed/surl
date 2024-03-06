![Go Build](https://github.com/jacobrreed/surl/actions/workflows/go.yml/badge.svg)

### URL Shortener in Go
This is a simple URL shortener API I wrote in Go as my first Go project. Comes with docker compose file to startup mongodb and mongo-express.

#### How to run
 - Clone the repository
 - Add .env file in directory and set `MONGO_PASSWORD="your_password"`
 - Run `docker-compose up -d` to start the mongodb and mongo-express
 - Run `go run main.go` to start the server
 - Go server listens on port 5000 and has the following routes:
  - `POST /shorten`
    - Request body: `{ "longUrl": "https://yourlongurlhere.com/123123123/asdasd/11"}`
    - Returns:
     ```json
     {
      "db_id": "65e7ce4213a5684b21b2abf6",
      "expires": "2024-03-10 21:00:34",
      "newUrl": "http://localhost:5000/woYWAKASR"
     }
    ```
 - `GET /:shortUrl`
    - Redirects to the original long url
