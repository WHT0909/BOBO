package main

import (
	"database/sql"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // 隐式导入驱动
)

var DB *sql.DB

func InitDB(baseDir string) {
	var err error
	dbPath := filepath.Join(baseDir, "bobo.db")
	// 打开（或创建）名为 bobo.db 的文件
	DB, err = sql.Open("sqlite3", dbPath)
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
		category TEXT,
		parent_id INTEGER DEFAULT 0
	);`

	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// 检查并添加 parent_id 列
	var hasColumn bool
	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('projects') WHERE name = 'parent_id'").Scan(&hasColumn)
	if err == nil && !hasColumn {
		DB.Exec("ALTER TABLE projects ADD COLUMN parent_id INTEGER DEFAULT 0")
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
