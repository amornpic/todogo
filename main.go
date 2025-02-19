package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Todo represents a todo item in our application
type Todo struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Global DB variable
var DB *gorm.DB

// ConnectDatabase establishes connection to PostgreSQL
func ConnectDatabase() {
	// Load the .env file
	errenv := godotenv.Load()
	if errenv != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the Todo model
	DB.AutoMigrate(&Todo{})
	fmt.Println("Database connected and migrated successfully")
}

// Create a new todo item
func CreateTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	DB.Create(&todo)
	return c.Status(fiber.StatusCreated).JSON(todo)
}

// Get all todos
func GetTodos(c *fiber.Ctx) error {
	var todos []Todo
	DB.Find(&todos)
	return c.JSON(todos)
}

// Get a single todo by ID
func GetTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	var todo Todo
	result := DB.First(&todo, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}
	return c.JSON(todo)
}

// Update a todo
func UpdateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	var todo Todo
	result := DB.First(&todo, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	// Parse the updated data
	updateData := new(Todo)
	if err := c.BodyParser(updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Update the todo
	if updateData.Title != "" {
		todo.Title = updateData.Title
	}
	if updateData.Description != "" {
		todo.Description = updateData.Description
	}
	todo.Completed = updateData.Completed

	DB.Save(&todo)
	return c.JSON(todo)
}

// Delete a todo
func DeleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	var todo Todo
	result := DB.First(&todo, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}
	DB.Delete(&todo)
	return c.SendStatus(fiber.StatusNoContent)
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")
	v1 := api.Group("/v1")
	todos := v1.Group("/todos")

	todos.Post("/", CreateTodo)
	todos.Get("/", GetTodos)
	todos.Get("/:id", GetTodo)
	todos.Put("/:id", UpdateTodo)
	todos.Delete("/:id", DeleteTodo)
}

func main() {
	// Connect to database
	ConnectDatabase()

	// Create Fiber instance
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	setupRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}
