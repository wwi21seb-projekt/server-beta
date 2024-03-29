package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/initializers"
	"github.com/wwi21seb-projekt/server-beta/internal/router"
	"github.com/wwi21seb-projekt/server-beta/internal/routines"
	"os"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	fmt.Println("Start server...")

	// Start daily routines
	go routines.StartDailyRoutines()

	// Define a port using argument flag
	// Set default port to :8080
	port := flag.String("port", "8080", "Port on which the server will run")
	flag.Parse()

	gin.SetMode(os.Getenv("GIN_MODE"))
	r := router.SetupRouter()
	err := r.Run(":" + *port)
	if err != nil {
		panic("Failed to start router")
	} else {
		fmt.Println("Router started...")
	}

	defer initializers.CloseDbConnection()
}
