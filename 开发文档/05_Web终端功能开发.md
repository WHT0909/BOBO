# Web终端功能开发文档

## 1. 设计思路

### 1.1 技术栈选择
- **前端**: xterm.js - 专业的终端模拟器库，提供完整的终端UI和交互功能
- **后端**: Go + Gin + gorilla/websocket - 实现WebSocket通信
- **操作系统**: 默认支持Windows系统(cmd)，同时预留Linux/macOS支持

### 1.2 架构设计

```
┌─────────────────┐         WebSocket          ┌─────────────────┐
│   浏览器        │ ◄──────────────────────► │   Go后端        │
│  (xterm.js)     │                            │  (websocket)    │
└─────────────────┘                            └────────┬────────┘
                                                     │
                                                     │ 启动子进程
                                                     │
                                              ┌──────▼──────┐
                                              │  系统终端   │
                                              │ (cmd/bash)  │
                                              └─────────────┘
```

### 1.3 功能特性
- 完整的终端UI体验
- 实时命令输入和输出
- 支持Windows cmd指令
- 终端大小自适应
- 历史命令记录
- 光标控制

---

## 2. 研发流程

### 2.1 阶段一：项目依赖准备
- 添加 gorilla/websocket 包
- 检查并安装 xterm.js CDN资源

### 2.2 阶段二：后端实现
- 实现WebSocket连接升级
- 创建终端会话管理
- 实现命令执行和IO管道
- 处理终端大小调整

### 2.3 阶段三：前端实现
- 集成xterm.js库
- 建立WebSocket连接
- 实现终端输入输出处理
- 添加UI样式

### 2.4 阶段四：测试验证
- 功能测试
- 稳定性测试
- 兼容性测试

---

## 3. 实现细节

### 3.1 后端实现 (main.go)
1. 添加WebSocket路由 `/ws/terminal`
2. 终端会话结构体管理
3. 使用 `exec.Cmd` 启动cmd进程
4. 通过管道连接stdin/stdout/stderr
5. 双向转发WebSocket消息

### 3.2 前端实现 (terminal.html)
1. 引入xterm.js和xterm-addon-fit
2. 创建Terminal实例
3. 建立WebSocket连接
4. 绑定终端输入到WebSocket发送
5. 接收WebSocket消息写入终端

### 3.3 样式实现 (style.css)
- 终端容器样式
- 适配xterm.js主题
- 响应式布局
