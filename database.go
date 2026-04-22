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

	// 创建 projects 表
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

	// 创建 versions 表
	sqlStmt2 := `
	CREATE TABLE IF NOT EXISTS versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER,
		update_time TEXT,
		description TEXT,
		FOREIGN KEY(project_id) REFERENCES projects(id)
	);`

	_, err = DB.Exec(sqlStmt2)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt2)
		return
	}
}
