package main

import (
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"time"
)

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set!", key)
	}
	return value
}

func setupRedis() *redis.Client {
	url := getEnv("REDIS_URL")
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Fatal("Failed to parse REDIS_URL")
	}

	log.Debugf("Connecting to redis on %s", opts.Addr)
	return redis.NewClient(opts)
}

func main() {
	log.Info("Starting...")

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(cache.New(cache.Config{
		Expiration: time.Minute * 15,
		KeyGenerator: func(ctx *fiber.Ctx) string {
			return ctx.Path() + ctx.Query("token")
		},
	}))

	rdb := setupRedis()
	service := newService(rdb, getOAuth2Config())
	service.registerEndpoints(app)

	port := getEnv("PORT")
	log.Infof("Listening on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to serve app: %s", err.Error())
	}
	if err := rdb.Close(); err != nil {
		log.Errorf("Failed to close redis connection: %s", err)
	}
}
