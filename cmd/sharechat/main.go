package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/http"
	"github.com/soggycactus/sharechat.dev/sharechat/postgres"
	"github.com/soggycactus/sharechat.dev/sharechat/redis"
)

func main() {
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASS")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbName := os.Getenv("POSTGRES_NAME")
	redisUser := os.Getenv("REDIS_USER")
	redisPass := os.Getenv("REDIS_PASS")
	redisHost := os.Getenv("REDIS_HOST")

	dbstring := fmt.Sprintf("user=%v dbname=%v password=%v host=%v sslmode=disable", dbUser, dbName, dbPass, dbHost)
	driver := "postgres"

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		log.Fatal(err)
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	roomRepo := postgres.NewRoomRepository(db, "postgres")
	messageRepo := postgres.NewMessageRepository(db, "postgres")
	memberRepo := postgres.NewMemberRepository(db, "postgres")
	controller := sharechat.NewController(sharechat.NewControllerInput{
		RoomRepo:    roomRepo,
		MessageRepo: messageRepo,
		MemberRepo:  memberRepo,
		Queue:       redis.NewQueue(redisHost, redisUser, redisPass),
	})

	server := http.NewServer(controller)

	log.Print("starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
