package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// get commandline arguments
	args := os.Args
	if len(args) != 3 {
		fmt.Printf("You must provide exactly 2 arguments: host and proxy port\n. Example: %s http://localhost:8001 9090\n", args[0])
		os.Exit(1)
	}

	host := args[1]
	proxyPort := args[2]

	app := fiber.New()
	app.Use(cors.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowHeaders:     "*",
	}))

	api := app.Group("/")

	api.Get("/*",
		func(ctx *fiber.Ctx) error {
			log.Printf("route GET request: %s", ctx.Path())

			agent := fiber.AcquireAgent()
			agent.Request().Header.SetMethod(http.MethodGet)
			for k, v := range ctx.GetReqHeaders() {
				log.Printf("adding header %s: %s", k, v)
				agent.Request().Header.Add(k, v)
			}

			agent.Request().SetBody(ctx.Body())
			agent.Request().SetRequestURI(host + ctx.Path())

			if err := agent.Parse(); err != nil {
				return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}

			statusCode, body, errs := agent.Bytes()
			if len(errs) > 0 {
				return ctx.Status(statusCode).JSON(errs)
			}

			log.Printf("response status: %d", statusCode)
			log.Printf("response body: %s", body)

			return ctx.Status(statusCode).Send(body)
		})

	api.Post("/*",
		func(ctx *fiber.Ctx) error {
			log.Printf("route POST request: %s", ctx.Path())
			log.Printf("request body: %s", ctx.Body())

			agent := fiber.AcquireAgent()
			agent.Request().Header.SetMethod(http.MethodPost)
			for k, v := range ctx.GetReqHeaders() {
				log.Printf("adding header %s: %s", k, v)
				agent.Request().Header.Add(k, v)
			}
			agent.Request().SetBody(ctx.Body())
			agent.Request().SetRequestURI(host + ctx.Path())

			err := agent.Parse()
			if err != nil {
				log.Printf("error : %s", err)
				return ctx.SendStatus(fiber.StatusInternalServerError)
			}

			statusCode, body, errs := agent.Bytes()
			if len(errs) > 0 {
				log.Printf("post request error %s", err)
				return ctx.Status(statusCode).JSON(errs)
			}

			log.Printf("POST response status: %d", statusCode)
			log.Printf("POST response body: %s", body)

			return ctx.Status(statusCode).Send(body)
		})

	log.Printf("Starting application...")

	if err := app.Listen(":" + proxyPort); err != nil {
		log.Fatalf("Error starting cors proxy: %s \n", err)
	}
}
