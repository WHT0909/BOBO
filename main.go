package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 检查命令行参数 (bobo dev)
	if len(os.Args) < 2 || os.Args[1] != "dev" {
		println("请使用 'bobo dev' 来启动服务")
		return
	}

	InitDB()
	// 确保程序退出前关闭数据库连接
	defer DB.Close()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	// 调用自己定义的路由配置函数
	setupRoutes(r)

	r.Run(":5050")
}

// 路由配置函数
func setupRoutes(r *gin.Engine) {
	// 首页展示项目
	r.GET("/", func(c *gin.Context) {
		// 获取所有项目用于侧边栏展示
		rows, err := DB.Query("SELECT id, name, path, command, note, category, parent_id FROM projects")
		if err != nil {
			c.String(http.StatusInternalServerError, "查询项目失败: "+err.Error())
			return
		}
		var projects []Project
		for rows.Next() {
			var p Project
			if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Command, &p.Note, &p.Category, &p.ParentID); err != nil {
				continue
			}
			projects = append(projects, p)
		}
		rows.Close()

		// 获取当前选中的项目 ID
		selectedID := c.Query("id")
		var currentProject *Project
		var versions []Version
		if selectedID != "" {
			for _, p := range projects {
				if fmt.Sprintf("%d", p.ID) == selectedID {
					currentProject = &p
					break
				}
			}
			// 获取该项目的版本迭代记录
			if currentProject != nil {
				versionRows, _ := DB.Query("SELECT id, project_id, update_time, description FROM versions WHERE project_id = ? ORDER BY id DESC", currentProject.ID)
				for versionRows.Next() {
					var v Version
					versionRows.Scan(&v.ID, &v.ProjectID, &v.UpdateTime, &v.Description)
					versions = append(versions, v)
				}
			}
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Projects":       projects,
			"CurrentProject": currentProject,
			"Versions":       versions,
		})
	})

	// 添加项目接口
	r.POST("/add", func(c *gin.Context) {
		name := c.PostForm("name")
		path := c.PostForm("path")
		cmd := c.PostForm("cmd")
		note := c.PostForm("note")
		category := c.PostForm("category")
		parentID := c.PostForm("parent_id") // 父项目ID

		// 转换 parent_id，默认为 0（顶级项目）
		var parentIDInt int
		if parentID != "" {
			fmt.Sscanf(parentID, "%d", &parentIDInt)
		}

		// 如果是子项目（parent_id > 0），验证分类必须与父项目一致
		if parentIDInt > 0 {
			var parentCategory string
			err := DB.QueryRow("SELECT category FROM projects WHERE id = ?", parentIDInt).Scan(&parentCategory)
			if err != nil {
				c.String(http.StatusNotFound, "找不到父项目")
				return
			}
			if category != parentCategory {
				c.String(http.StatusBadRequest, fmt.Sprintf("子项目的分类必须与父项目一致。父项目分类为: %s", parentCategory))
				return
			}
		}

		DB.Exec("INSERT INTO projects (name, path, command, note, category, parent_id) VALUES (?, ?, ?, ?, ?, ?)",
			name, path, cmd, note, category, parentIDInt)
		c.Redirect(http.StatusSeeOther, "/")
	})

	// 核心启动接口
	r.GET("/run/:id", func(c *gin.Context) {
		id := c.Param("id")
		var p Project
		err := DB.QueryRow("SELECT path, command FROM projects WHERE id = ?", id).Scan(&p.Path, &p.Command)
		if err != nil {
			c.String(http.StatusNotFound, "找不到该项目")
			return
		}

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", p.Command)
		} else {
			cmd = exec.Command("sh", "-c", p.Command)
		}
		cmd.Dir = p.Path

		err = cmd.Start() // 异步启动
		if err != nil {
			c.String(http.StatusInternalServerError, "启动失败: "+err.Error())
			return
		}
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// 删除项目接口
	r.GET("/delete/:id", func(c *gin.Context) {
		id := c.Param("id")
		_, err := DB.Exec("DELETE FROM projects WHERE id = ?", id)
		if err != nil {
			c.String(http.StatusInternalServerError, "删除失败: "+err.Error())
			return
		}
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// 修改项目接口
	r.GET("/edit/:id", func(c *gin.Context) {
		id := c.Param("id")
		var p Project
		err := DB.QueryRow("SELECT id, name, path, command, note, category, parent_id FROM projects WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Path, &p.Command, &p.Note, &p.Category, &p.ParentID)
		if err != nil {
			c.String(http.StatusNotFound, "找不到该项目")
			return
		}

		// 获取所有项目用于侧边栏展示
		rows, err := DB.Query("SELECT id, name, path, command, note, category, parent_id FROM projects")
		if err != nil {
			c.String(http.StatusInternalServerError, "查询项目列表失败: "+err.Error())
			return
		}
		var projects []Project
		for rows.Next() {
			var proj Project
			if err := rows.Scan(&proj.ID, &proj.Name, &proj.Path, &proj.Command, &proj.Note, &proj.Category, &proj.ParentID); err != nil {
				continue
			}
			projects = append(projects, proj)
		}
		rows.Close()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Projects":       projects,
			"CurrentProject": p,
			"EditMode":       true,
		})
	})

	// 保存修改项目接口
	r.POST("/update/:id", func(c *gin.Context) {
		id := c.Param("id")
		name := c.PostForm("name")
		path := c.PostForm("path")
		cmd := c.PostForm("cmd")
		note := c.PostForm("note")
		category := c.PostForm("category")
		parentID := c.PostForm("parent_id")

		var parentIDInt int
		if parentID != "" {
			fmt.Sscanf(parentID, "%d", &parentIDInt)
		}

		// 如果是子项目（parent_id > 0），验证分类必须与父项目一致
		if parentIDInt > 0 {
			var parentCategory string
			err := DB.QueryRow("SELECT category FROM projects WHERE id = ?", parentIDInt).Scan(&parentCategory)
			if err != nil {
				c.String(http.StatusNotFound, "找不到父项目: "+err.Error())
				return
			}
			if category != parentCategory {
				c.String(http.StatusBadRequest, fmt.Sprintf("子项目的分类必须与父项目一致。父项目分类为: %s", parentCategory))
				return
			}
		}

		_, err := DB.Exec("UPDATE projects SET name = ?, path = ?, command = ?, note = ?, category = ?, parent_id = ? WHERE id = ?", name, path, cmd, note, category, parentIDInt, id)
		if err != nil {
			c.String(http.StatusInternalServerError, "更新失败: "+err.Error())
			return
		}
		c.Redirect(http.StatusSeeOther, "/?id="+id)
	})

	// 添加版本迭代接口
	r.POST("/add-version/:id", func(c *gin.Context) {
		id := c.Param("id")
		updateTime := c.PostForm("update_time")
		description := c.PostForm("description")

		// 如果用户没有提供时间，则使用当前时间
		if updateTime == "" {
			updateTime = time.Now().Format("2006-01-02 15:04:05")
		}

		// 检查数据库连接并执行插入
		_, err := DB.Exec("INSERT INTO versions (project_id, update_time, description) VALUES (?, ?, ?)", id, updateTime, description)
		if err != nil {
			c.String(http.StatusInternalServerError, "添加版本迭代失败: "+err.Error())
			return
		}
		c.Redirect(http.StatusSeeOther, "/?id="+id)
	})

	// 处理直接访问 /add-version/:id 的 GET 请求，重定向回项目详情
	r.GET("/add-version/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(http.StatusSeeOther, "/?id="+id)
	})

	// 删除版本迭代记录
	r.GET("/delete-version/:versionId", func(c *gin.Context) {
		versionId := c.Param("versionId")

		// 先获取该版本的 project_id 用于重定向
		var projectId int
		err := DB.QueryRow("SELECT project_id FROM versions WHERE id = ?", versionId).Scan(&projectId)
		if err != nil {
			c.String(http.StatusNotFound, "找不到该版本记录")
			return
		}

		// 删除版本记录
		_, err = DB.Exec("DELETE FROM versions WHERE id = ?", versionId)
		if err != nil {
			c.String(http.StatusInternalServerError, "删除失败: "+err.Error())
			return
		}

		// 重定向回项目详情页
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?id=%d", projectId))
	})

	// 编辑版本迭代 - 显示编辑表单
	r.GET("/edit-version/:versionId", func(c *gin.Context) {
		versionId := c.Param("versionId")

		// 获取版本信息
		var v Version
		err := DB.QueryRow("SELECT id, project_id, update_time, description FROM versions WHERE id = ?", versionId).Scan(&v.ID, &v.ProjectID, &v.UpdateTime, &v.Description)
		if err != nil {
			c.String(http.StatusNotFound, "找不到该版本记录: "+err.Error())
			return
		}

		// 获取所有项目用于侧边栏展示
		rows, err := DB.Query("SELECT id, name, path, command, note, category, parent_id FROM projects")
		if err != nil {
			c.String(http.StatusInternalServerError, "查询项目列表失败: "+err.Error())
			return
		}
		var projects []Project
		for rows.Next() {
			var p Project
			if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Command, &p.Note, &p.Category, &p.ParentID); err != nil {
				continue
			}
			projects = append(projects, p)
		}
		rows.Close()

		// 获取当前项目的所有版本记录
		versionRows, err := DB.Query("SELECT id, project_id, update_time, description FROM versions WHERE project_id = ? ORDER BY id DESC", v.ProjectID)
		if err != nil {
			c.String(http.StatusInternalServerError, "查询版本记录失败: "+err.Error())
			return
		}
		var versions []Version
		for versionRows.Next() {
			var version Version
			if err := versionRows.Scan(&version.ID, &version.ProjectID, &version.UpdateTime, &version.Description); err != nil {
				continue
			}
			versions = append(versions, version)
		}
		versionRows.Close()

		// 找到当前项目
		var currentProject *Project
		for _, p := range projects {
			if p.ID == v.ProjectID {
				currentProject = &p
				break
			}
		}

		// 如果找不到对应的项目，返回错误
		if currentProject == nil {
			c.String(http.StatusNotFound, "找不到对应的项目")
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Projects":        projects,
			"CurrentProject":  currentProject,
			"Versions":        versions,
			"EditVersion":     v,
			"EditVersionMode": true,
		})
	})

	// 更新版本迭代记录
	r.POST("/update-version/:versionId", func(c *gin.Context) {
		versionId := c.Param("versionId")
		updateTime := c.PostForm("update_time")
		description := c.PostForm("description")

		// 验证必填字段
		if updateTime == "" || description == "" {
			c.String(http.StatusBadRequest, "更新时间和描述不能为空")
			return
		}

		// 获取 project_id 用于重定向
		var projectId int
		err := DB.QueryRow("SELECT project_id FROM versions WHERE id = ?", versionId).Scan(&projectId)
		if err != nil {
			c.String(http.StatusNotFound, "找不到该版本记录: "+err.Error())
			return
		}

		// 更新版本记录
		result, err := DB.Exec("UPDATE versions SET update_time = ?, description = ? WHERE id = ?", updateTime, description, versionId)
		if err != nil {
			c.String(http.StatusInternalServerError, "更新失败: "+err.Error())
			return
		}

		// 检查是否真的更新了记录
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.String(http.StatusNotFound, "没有找到要更新的记录")
			return
		}

		// 重定向回项目详情页
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?id=%d", projectId))
	})
}
