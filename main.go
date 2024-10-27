package main

import (
	"my-api/board"
	"my-api/database"
	"my-api/server"
	"net/http"
)

func main() {

	app := server.New("0.0.0.0:9000")

	serverLogger := server.GetLogger()
	app.Context["logger"] = serverLogger

	serverDatabase := database.New("user:password@tcp(172.25.20.120:18204)/DEV?charset=utf8&parseTime=true", serverLogger)
	// Table 마이그레이션 -> TODO entity 폴더 생성 ..
	session := serverDatabase.GetSession()
	session.AutoMigrate(&board.BoardInfo{})
	app.Context["database"] = serverDatabase

	// root path
	app.AddRoute("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World"))
	})

	// routers
	board := board.New(&app.Context)
	app.AddRouteByRouter(board)

	// Server 실행
	app.Run()
}
