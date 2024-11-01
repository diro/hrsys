package main

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql" // 使用 MySQL 驅動
    "log"
    "net/http"
)

func main() {
    db, err := sql.Open("mysql", "username:password@tcp(your-rds-endpoint:3306)/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    log.Println("Server is running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

