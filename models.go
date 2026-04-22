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