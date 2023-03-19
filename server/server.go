package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"github.com/rkonfj/hu60bot/convo"
	"github.com/rkonfj/hu60bot/pkg/hu60"
	"github.com/sirupsen/logrus"
)

type WebsocketManager struct {
	upgrader         websocket.Upgrader
	hu60Client       *hu60.Client
	websocketConnMap map[int]*websocket.Conn
	cm               *convo.ConversationManager
	options          ServerOptions
}

type Hu60Msg struct {
	ID         int       `json:"id"`
	ToUID      int       `json:"touid"`
	ByUID      int       `json:"byuid"`
	Type       int       `json:"type"`
	Read       int       `json:"isread"`
	Content    string    `json:"content"`
	CreateTime time.Time `json:"ctime"`
}

type BotEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type BotCmd struct {
	Action string `json:"action"`
	Data   any    `json:"data"`
}

type ChatResponse struct {
	NewConversation bool   `json:"newConversation"`
	Response        string `json:"response"`
}

func NewWebsocketManager(opts ServerOptions, cm *convo.ConversationManager) *WebsocketManager {
	return &WebsocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		hu60Client:       hu60.NewClient(opts.Hu60wap6APIURL),
		websocketConnMap: make(map[int]*websocket.Conn),
		cm:               cm,
		options:          opts,
	}
}

func (m *WebsocketManager) Push(msg *Hu60Msg) error {
	if ws, ok := m.websocketConnMap[msg.ToUID]; ok {
		err := ws.WriteJSON(BotEvent{Event: "msg", Data: msg})
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("uid %d not online", msg.ToUID)
	}
}

func getRequestParam(r *http.Request, name string, noCookie bool) (value string) {
	value = r.PostFormValue("_" + name) // POST
	if value == "" {
		value = r.FormValue("_" + name) // GET
	}
	if value == "" {
		value = r.Header.Get("x-" + name) // Header
	}
	if value == "" && !noCookie {
		cookie, err := r.Cookie("hu60_" + name) // Cookie
		if err == nil {
			value = cookie.Value
		}
	}
	return
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

// checkSameOrigin returns true if the origin is not set or is equal to the request host.
func checkSameOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	return equalASCIIFold(u.Host, r.Host)
}

func (m *WebsocketManager) Run() error {
	http.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		noCookie := false
		var header http.Header = make(http.Header)

		// 支持跨域访问
		if !checkSameOrigin(r) {
			noCookie = true
		}
		origin := getRequestParam(r, "origin", true)
		if origin != "" {
			noCookie = true
			header["Access-Control-Allow-Origin"] = []string{origin}
		}

		ws, err := m.upgrader.Upgrade(w, r, header)
		if err != nil {
			logrus.Error("ws upgrade error: ", err.Error())
			return
		}

		// 获取sid（跨域时禁用cookie）
		sid := getRequestParam(r, "sid", noCookie)
		if sid == "" {
			m.responseUnauthenticated(ws)
			logrus.Warn("unauthenticated: ", err.Error())
			return
		}

		res, err := m.hu60Client.GetProfile(context.Background(), sid)
		if err != nil {
			m.responseUnauthenticated(ws)
			logrus.Warn("unauthenticated: ", err.Error())
			return
		}

		err = ws.WriteMessage(websocket.TextMessage, []byte(`{"event": "connected"}`))
		if err != nil {
			logrus.Error(err)
			return
		}
		m.websocketConnMap[res.Uid] = ws
		logrus.Info("user ", res.Name, " is connected")

		go func(userProfile hu60.GetProfileResponse, ws *websocket.Conn) {
			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					m.websocketConnMap[userProfile.Uid].Close()
					delete(m.websocketConnMap, userProfile.Uid)
					logrus.Info("user ", userProfile.Name, " is disconnected")
					break
				}
				var cmd BotCmd
				err = json.NewDecoder(strings.NewReader(string(msg))).Decode(&cmd)
				if err != nil {
					m.responseError(ws, err)
					return
				}
				m.processBotAction(cmd, res.Uid, ws)
			}
		}(res, ws)
	})
	logrus.Info("bot listening on ", m.options.Listen, " for interact now. websocket endpoint is /v1/ws")
	return http.ListenAndServe(m.options.Listen, nil)
}

func (m *WebsocketManager) responseUnauthenticated(ws *websocket.Conn) {
	ws.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Unauthenticated"))
	ws.Close()
}

func (m *WebsocketManager) responseError(ws *websocket.Conn, err error) {
	e := ws.WriteJSON(BotEvent{Event: "error", Data: err.Error()})
	if e != nil {
		logrus.Error(e.Error())
	}
}

func (m *WebsocketManager) processBotAction(cmd BotCmd, uid int, ws *websocket.Conn) {
	if cmd.Action == "chat" {
		conversationKey := fmt.Sprintf("%d", uid)
		if d, ok := cmd.Data.(string); ok {
			responseText, newConversation, err := m.cm.Ask(d, conversationKey)
			cr := ChatResponse{
				NewConversation: newConversation,
				Response:        responseText,
			}
			if err != nil {
				cr = ChatResponse{
					NewConversation: true,
					Response:        err.Error(),
				}
			}
			ws.WriteJSON(BotEvent{Event: "chat", Data: cr})
		}
		return
	}
	if cmd.Action == "rmconvo" {
		conversationKey := fmt.Sprintf("%d", uid)
		m.cm.MarkExpired(conversationKey)
		ws.WriteJSON(BotEvent{Event: "rmconvo", Data: "ok"})
		return
	}
	if cmd.Action == "ping" {
		ws.WriteJSON(BotEvent{Event: "ping", Data: "pong"})
		return
	}
	m.responseError(ws, errors.New("unsupported action"))
}
