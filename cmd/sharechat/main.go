package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/pressly/goose/v3"
	"github.com/rs/cors"
	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/http"
	"github.com/soggycactus/sharechat.dev/sharechat/postgres"
	"github.com/soggycactus/sharechat.dev/sharechat/redis"
)

var (
	upgrader       websocket.Upgrader = websocket.Upgrader{HandshakeTimeout: 5 * time.Second}
	allowedOrigins http.AllowedOrigins
)

func main() {
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASS")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbName := os.Getenv("POSTGRES_NAME")
	dbPort := os.Getenv("POSTGRES_PORT")
	redisUser := os.Getenv("REDIS_USER")
	redisPass := os.Getenv("REDIS_PASS")
	redisHost := os.Getenv("REDIS_HOST")

	flag.Var(&allowedOrigins, "allowed-origin", "allowed origins for CORS")
	flag.Parse()

	log.Printf("allowed origins: %v", allowedOrigins)

	dbstring := fmt.Sprintf("user=%v dbname=%v password=%v host=%v port=%v sslmode=disable", dbUser, dbName, dbPass, dbHost, dbPort)
	driver := "postgres"

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		log.Fatal(err)
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redisv8.NewClient(&redisv8.Options{
		Addr:     redisHost,
		Username: redisUser,
		Password: redisPass,
	})

	roomRepo := postgres.NewRoomRepository(db, "postgres")
	messageRepo := postgres.NewMessageRepository(db, "postgres")
	memberRepo := postgres.NewMemberRepository(db, "postgres")
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       redis.NewQueue(*redisClient),
		Healthcheck: func(c context.Context) error {
			if err := db.PingContext(c); err != nil {
				return err
			}

			if err := redisClient.Ping(c).Err(); err != nil {
				return err
			}

			return nil
		},
	})

	server := http.NewServer(controller, upgrader, cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "HEAD", "POST", "OPTIONS"},
	})

	log.Print("starting sharechat server on port 8080")
	log.Fatal(server.Server.ListenAndServe())
}
