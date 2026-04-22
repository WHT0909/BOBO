## 🚀 Bobo Dev 项目开发技术文档

### 1. 项目概览
* **目标**：创建一个轻量级、单文件运行的项目管理工具。
* **核心功能**：
    * 通过 `bobo dev` 启动 Web 面板。
    * 在面板中展示、添加、备注本地项目。
    * 在面板中一键运行项目（如启动 Python/Node 服务）。
* **技术栈**：Go, Gin (Web 框架), SQLite (数据存储), Go Templates (前端渲染)。

---

### 2. 环境准备 (第一步)
在开始写代码前，请确保你的电脑已安装以下工具：
1.  **Go 编译器**：前往 [golang.google.cn](https://golang.google.cn/) 下载并安装（建议版本 1.20+）。
2.  **C 编译器**（由于使用 SQLite）：
    * Windows 用户建议安装 **TDM-GCC** 或 **Mingw-w64**。
    * macOS 用户自带（需安装命令行开发工具）。
3.  **VS Code**：安装 **Go 扩展插件**。

---

### 3. 开发流程建议 (优先级顺序)

按照以下顺序开发，可以保证你每完成一步都能看到反馈：
1.  **阶段一：项目初始化**（建立文件夹和依赖）。
2.  **阶段二：数据库设计**（定义如何存储项目信息）。
3.  **阶段三：后端接口 (API)**（实现增删改查逻辑）。
4.  **阶段四：前端 UI (HTML 模板)**（让面板能看、能点）。
5.  **阶段五：核心功能——启动项目**（利用 Go 调用系统命令）。
6.  **阶段六：命令行封装**（实现 `bobo dev` 指令）。

---

### 4. 详细步骤指南

#### 第一步：初始化项目
在终端执行以下命令：
```bash
mkdir bobo-dev
cd bobo-dev
go mod init bobo-dev
# 下载依赖
go get -u github.com/gin-gonic/gin
go get -u github.com/mattn/go-sqlite3
```

#### 第二步：设计目录结构
推荐的结构如下，清晰且符合 Go 习惯：
```text
bobo-dev/
├── main.go           # 程序入口，处理命令行参数
├── database.go       # 数据库初始化与操作
├── handlers.go       # Web 路由的处理函数
├── models.go         # 数据模型定义
├── templates/        # 存放 HTML 文件
│   └── index.html
└── static/           # 存放 CSS/JS (可选)
```

#### 第三步：数据模型 (models.go)
定义你的项目包含哪些信息：
```go
package main

type Project struct {
    ID      int    `json:"id"`
    Name    string `json:"name"`
    Path    string `json:"path"`
    Command string `json:"command"` // 启动命令，如 python app.py
    Note    string `json:"note"`
    Status  string `json:"status"`  // 运行中 / 已停止
}
```

#### 第四步：数据库初始化 (database.go)
使用 SQLite 存储数据，即便程序关闭，项目信息也不会丢失。


#### 第五步：核心功能——执行外部命令
这是项目最酷的地方。你需要用到 Go 的 `os/exec` 包。
* **逻辑**：当用户在网页点击“启动”时，后端获取该项目的 `Path` 和 `Command`。
* **代码思路**：
    ```go
    cmd := exec.Command("sh", "-c", project.Command) // Linux/macOS
    // Windows 下使用: exec.Command("cmd", "/C", project.Command)
    cmd.Dir = project.Path
    err := cmd.Start() // 异步启动，不阻塞 Web 服务
    ```

#### 第六步：前端页面 (templates/index.html)
利用 **Go 模板** 循环显示项目：
```html
{{ range .projects }}
<div class="project-card">
    <h3>{{ .Name }}</h3>
    <p>位置: {{ .Path }}</p>
    <button onclick="startProject({{.ID}})">启动项目</button>
</div>
{{ end }}
```

---

### 5. 你需要实现的功能清单 (Checklist)

* [ ] **初始化数据库**：程序启动时检查 `bobo.db` 是否存在，不存在则创建表。
* [ ] **添加项目 API**：接收前端传来的路径和名称，存入 SQLite。
* [ ] **项目列表 API**：从 SQLite 读取所有项目，通过 `c.HTML` 渲染到页面。
* [ ] **运行逻辑**：实现一个接口，点击后调用 `exec.Command`。
* [ ] **命令行包装**：在 `main.go` 中解析参数，如果输入 `dev` 则启动 Gin 服务并打开浏览器。

---

### 6. 给你的开发小贴士

1.  **如何自动打开浏览器？**
    使用第三方库 `github.com/pkg/browser`，在 `r.Run(":5051")` 之前调用 `browser.OpenURL("http://localhost:5051")`。
2.  **前端美化：**
    不需要学复杂的 CSS。在 HTML 头部加入这一行，你的页面会立刻变高级：
    `<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@exampledev/new.css@1.1.2/new.min.css">`（或者使用 **Pico.css**）。
3.  **调试技巧：**
    先用 `go run .` 运行。当你觉得功能完美了，再用 `go build -o bobo` 编译成可执行文件。

**建议：** 你可以先尝试写出 `main.go` 的基础框架，让它能跑起来并在 5051 端口显示一个 "Hello Bobo" 的 HTML 页面。如果你准备好了，我可以帮你写出这部分的代码原型。