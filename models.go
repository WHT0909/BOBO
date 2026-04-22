package main

// Project 结构体对应数据库里的一行数据
type Project struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Command  string `json:"command"`
	Note     string `json:"note"`
	Category string `json:"category"`
}

// Version 结构体对应版本迭代记录
type Version struct {
	ID          int    `json:"id"`
	ProjectID   int    `json:"project_id"`
	UpdateTime  string `json:"update_time"`
	Description string `json:"description"`
}