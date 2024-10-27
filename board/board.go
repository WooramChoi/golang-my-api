package board

import (
	"encoding/json"
	"fmt"
	"log"
	"my-api/database"
	"my-api/server"
	"my-api/utils"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Board struct {
	Router         map[string]server.RouteHandler
	serverDatabase *database.Database
	serverLogger   *log.Logger
}

type BoardInfo struct {
	gorm.Model
	Title     string `gorm:"size:255" json:"title"`
	Content   string `gorm:"size:4000" json:"content"`
	PlainText string `gorm:"type:text" json:"plain_text"`
	YnUse     string `gorm:"size:1;default:Y" json:"yn_use"`
	Name      string `gorm:"size:50" json:"name"`
	Pwd       string `gorm:"size:255" json:"pwd"`
}

type BoardCommon struct {
	// gorm.Model
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`

	Title string `json:"title"`
	YnUse string `json:"yn_use"`
	Name  string `json:"name"`
}

type BoardSummary struct {
	BoardCommon
	ContentSummary string `json:"content_summary"`
}

type BoardDetails struct {
	BoardCommon
	Content   string `json:"content"`
	PlainText string `json:"plain_text"`
}

func New(context *map[string]interface{}) *Board {
	board := Board{}
	board.Router = make(map[string]server.RouteHandler)

	serverDatabase, ok := (*context)["database"].(*database.Database)
	if ok {
		board.serverDatabase = serverDatabase
	}

	serverLogger, ok := (*context)["logger"].(*log.Logger)
	if ok {
		board.serverLogger = serverLogger
	}

	board.Router["/boards"] = board.boards
	board.Router["/boards/{ID}"] = board.boardsById
	return &board
}

// my-api/server Router interface
func (board *Board) GetRouter() map[string]server.RouteHandler {
	return board.Router
}

// [*] /boards
func (board *Board) boards(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		board.getBoards(w, req)
	case http.MethodPost:
		board.postBoards(w, req)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// [GET] /boards
func (board *Board) getBoards(w http.ResponseWriter, req *http.Request) {

	session := board.serverDatabase.GetSession()
	session = session.Scopes(database.Paginate(req))

	for column, value := range req.URL.Query() {
		// TODO column-value 유효성 확인?
		session = session.Scopes(database.WhereEqual(column, value[0])) // TODO value list 처리
	}
	var boardInfos []BoardInfo
	if err := session.Find(&boardInfos).Error; err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	listBoardSummary := []BoardSummary{}
	for _, info := range boardInfos {
		// TODO struct to struct 속성 복사: struct 임베디드 고려
		boardSummary := BoardSummary{
			BoardCommon: BoardCommon{
				ID:        info.ID,
				CreatedAt: info.CreatedAt,
				UpdatedAt: info.UpdatedAt,
				DeletedAt: info.DeletedAt.Time,
				Title:     info.Title,
				YnUse:     info.YnUse,
				Name:      info.Name,
			},
			ContentSummary: utils.Substr(info.PlainText, 255),
		}
		listBoardSummary = append(listBoardSummary, boardSummary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listBoardSummary)
}

// [POST] /boards
func (board *Board) postBoards(w http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content-type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	var info BoardInfo
	session := board.serverDatabase.GetSession()
	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := session.Create(&info).Error; err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// http.Redirect(w, req, fmt.Sprintf("/boards/%d", info.ID), http.StatusMovedPermanently)
	http.Redirect(w, req, fmt.Sprintf("/boards/%d", info.ID), http.StatusCreated)
}

// [*] /boards/{ID}
func (board *Board) boardsById(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		board.getBoardsById(w, req)
	case http.MethodPatch:
		board.patchBoardsById(w, req)
	case http.MethodDelete:
		board.deleteBoardsById(w, req)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// [GET] /boards/{ID}
func (board *Board) getBoardsById(w http.ResponseWriter, req *http.Request) {

	var info BoardInfo

	session := board.serverDatabase.GetSession()

	if err := session.First(&info, req.PathValue("ID")).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// TODO struct to struct 속성 복사: struct 임베디드 고려
	boardDetails := BoardDetails{
		BoardCommon: BoardCommon{
			ID:        info.ID,
			CreatedAt: info.CreatedAt,
			UpdatedAt: info.UpdatedAt,
			DeletedAt: info.DeletedAt.Time,
			Title:     info.Title,
			YnUse:     info.YnUse,
			Name:      info.Name,
		},
		Content:   info.Content,
		PlainText: info.PlainText,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boardDetails)
}

// [PATCH] /boards/{ID}
func (board *Board) patchBoardsById(w http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content-type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	var info BoardInfo

	session := board.serverDatabase.GetSession()

	if err := session.First(&info, req.PathValue("ID")).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session.Save(&info)
	http.Redirect(w, req, fmt.Sprintf("/boards/%d", info.ID), http.StatusNoContent)
}

// [DELETE] /boards/{ID}
func (board *Board) deleteBoardsById(w http.ResponseWriter, req *http.Request) {

	var info BoardInfo

	session := board.serverDatabase.GetSession()

	if err := session.First(&info, req.PathValue("ID")).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	session.Delete(&info)
	w.WriteHeader(http.StatusNoContent)
}
