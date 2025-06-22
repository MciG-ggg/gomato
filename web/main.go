// web/main.go
package main

import (
	"gomato/task" // 导入我们项目中的task包
	"log"         // 用于打印日志
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket" // 导入websocket包
)

// WebSocket upgrader，用于将HTTP连接升级为WebSocket连接
var upgrader = websocket.Upgrader{
	// CheckOrigin是一个函数，用于检查请求的来源是否允许。
	// 这里我们允许所有来源，但在生产环境中应该进行更严格的检查。
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// hub 管理所有的客户端连接和广播消息
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	timer      *time.Ticker
	timerState string // "running", "paused"
	timeLeft   time.Duration
}

// newHub 创建一个新的Hub
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		timer:      time.NewTicker(1 * time.Second),
		timerState: "paused",
		timeLeft:   25 * time.Minute,
	}
}

// run 启动hub，监听各种channel
func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
			}
		case <-h.timer.C:
			if h.timerState == "running" && h.timeLeft > 0 {
				h.timeLeft -= 1 * time.Second
				message := []byte(h.timeLeft.String())
				for conn := range h.clients {
					conn.WriteMessage(websocket.TextMessage, message)
				}
			}
		}
	}
}

// handleWebSocket 处理WebSocket请求
func handleWebSocket(c *gin.Context, hub *Hub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.Close()

	hub.register <- conn

	for {
		// 读取客户端消息 (start, pause, reset)
		_, msg, err := conn.ReadMessage()
		if err != nil {
			hub.unregister <- conn
			break
		}

		command := string(msg)
		switch command {
		case "start":
			hub.timerState = "running"
		case "pause":
			hub.timerState = "paused"
		case "reset":
			hub.timerState = "paused"
			hub.timeLeft = 25 * time.Minute
			// 立即广播重置后的时间
			message := []byte(hub.timeLeft.String())
			for client := range hub.clients {
				client.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}

func main() {
	// 创建一个默认的Gin引擎。
	// 它包含了Logger和Recovery中间件，有助于开发和调试。
	router := gin.Default()

	// 初始化任务管理器
	// 这个实例会在整个应用生命周期中存在，用于管理所有任务。
	taskManager := task.NewTaskManager()

	// 加载静态文件，例如CSS和JavaScript。
	// 第一个参数是URL中的路径，第二个参数是文件系统中的路径。
	// 这样，当浏览器请求 /static/style.css 时，Gin会去 web/static/ 目录寻找 style.css 文件。
	router.Static("/static", "./web/static")

	// 加载HTML模板。
	// Gin可以渲染HTML模板，这对于构建动态网页非常有用。
	router.LoadHTMLGlob("web/templates/*")

	// 定义一个根路由 ("/") 的处理器。
	// 当用户访问网站主页时，这个函数会被调用。
	router.GET("/", func(c *gin.Context) {
		// 渲染 "index.html" 模板。
		// http.StatusOK 表示HTTP状态码 200 (成功)。
		// gin.H 是一个map[string]interface{}的快捷方式，用于向模板传递数据。
		// 这里我们传递了一个标题 "Go番茄钟Web应用"。
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Go番茄钟Web应用",
		})
	})

	hub := newHub()
	go hub.run()

	// 为WebSocket添加一个新的路由
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c, hub)
	})

	// ---- API 路由 ----
	// 创建一个API路由组，方便管理
	api := router.Group("/api")
	{
		// 获取所有任务
		// GET /api/tasks
		api.GET("/tasks", func(c *gin.Context) {
			// 调用taskManager获取所有任务
			tasks := taskManager.ListTasks()
			// 以JSON格式返回任务列表
			c.JSON(http.StatusOK, tasks)
		})

		// 添加一个新任务
		// POST /api/tasks
		api.POST("/tasks", func(c *gin.Context) {
			// 定义一个临时的struct来绑定请求的JSON数据
			var newTask struct {
				Description string `json:"description" binding:"required"`
			}

			// 将请求体中的JSON绑定到newTask变量上
			// 如果绑定失败（例如，description字段缺失），则返回一个错误。
			if err := c.ShouldBindJSON(&newTask); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "任务描述不能为空"})
				return
			}

			// 调用taskManager添加新任务
			addedTask := taskManager.AddTask(newTask.Description)
			// 返回新创建的任务，HTTP状态码 201 Created
			c.JSON(http.StatusCreated, addedTask)
		})
	}

	// 启动HTTP服务器，并监听在8080端口。
	// router.Run() 会阻塞当前goroutine，直到程序被中断。
	router.Run(":8080")
}
