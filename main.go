package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Todos struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func main() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		panic(err)
	}
	debug.PrintStack()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, status TEXT)`)
	if err != nil {
		debug.PrintStack()
		panic(err)
	}
	defer db.Close()
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(static.Serve("/", static.LocalFile("./dist", true)))
	// r.StaticFS("/", http.Dir("dist"))

	// r.GET("/", func(c *gin.Context) {
	// 	c.HTML(200, "index.html", nil)
	//   })
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/view-todos", func(c *gin.Context) {
		var todos []Todos
		// var version string
		rows, err := db.Query("SELECT id, title, status FROM todos")
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			var id int
			var title string
			var status string
			err = rows.Scan(&id, &title, &status)
			if err != nil {
				panic(err)
			}
			fmt.Println(id, title, status)
			todos = append(todos, Todos{id, title, status})
		}
		// jsonStr, err := json.Marshal(todos)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": todos,
		})
	})
	r.POST("/add-todos", createTodo)
	r.Run("0.0.0.0:5001") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func createTodo(c *gin.Context) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		panic(err)
	}
	var todo Todos
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the todo into the database
	result, err := db.Exec("INSERT INTO todos (title, status) VALUES (?, ?)", todo.Title, todo.Status)
	if err != nil {
		fmt.Println("Error inserting todo:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	// Get the ID of the inserted todo
	id, _ := result.LastInsertId()
	todo.ID = int(id)

	c.JSON(http.StatusCreated, todo)
}
