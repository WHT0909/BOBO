package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // 隐式导入驱动
)

var DB *sql.DB

func InitDB() {
	var err error
	// 打开（或创建）名为 bobo.db 的文件
	DB, err = sql.Open("sqlite3", "./bobo.db")
	if err != nil {
		log.Fatal(err)
	}

	// 创建表的 SQL 语句
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		path TEXT,
		command TEXT,
		note TEXT,
		category TEXT
	);`

	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}
