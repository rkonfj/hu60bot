package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"github.com/rkonfj/hu60bot/pkg/hu60"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type WebsocketManager struct {
	upgrader               websocket.Upgrader
	hu60Client             *hu60.Client
	connMap                map[int][]*websocket.Conn
	connMapUpdateLock      sync.Mutex
	options                ServerOptions
	depOkSig               chan int
	broadcastEventChan     chan BotEvent
	connEventChan          chan ConnBotEvent
	userEventChan          chan UserBotEvent
	unsubscribedEvents     map[int][]string
	unsubscribedEventsLock sync.Mutex
	robotActionHub         map[int][]string
	robotActionHubLock     sync.Mutex
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
	ID     string `json:"id"`
	Action string `json:"action"`
	Data   any    `json:"data"`
}

type ConnBotEvent struct {
	Event BotEvent
	Uid   int
	Ws    *websocket.Conn
}

type UserBotEvent struct {
	Event BotEvent
	Uid   int
}

type UserOnlineStatus struct {
	UID   int `json:"uid"`
	Count int `json:"count"`
}

type ChatResponse struct {
	NewConversation bool   `json:"newConversation"`
	Response        string `json:"response"`
}

func NewWebsocketManager(opts ServerOptions) *WebsocketManager {
	return &WebsocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		hu60Client: hu60.NewClientWithConfig(hu60.Config{
			ApiURL:     opts.Hu60wap6APIURL,
			HTTPClient: http.DefaultClient,
			XFFHeader:  opts.BotXFF,
		}),
		connMap:                make(map[int][]*websocket.Conn),
		connMapUpdateLock:      sync.Mutex{},
		options:                opts,
		depOkSig:               make(chan int),
		broadcastEventChan:     make(chan BotEvent, 1024),
		connEventChan:          make(chan ConnBotEvent, 1024),
		userEventChan:          make(chan UserBotEvent, 1024),
		unsubscribedEvents:     make(map[int][]string),
		unsubscribedEventsLock: sync.Mutex{},
		robotActionHub:         make(map[int][]string),
		robotActionHubLock:     sync.Mutex{},
	}
}

func (m *WebsocketManager) Push(msg *Hu60Msg) {
	m.userEventChan <- UserBotEvent{
		Event: BotEvent{Event: "msg", Data: msg},
		Uid:   msg.ToUID,
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

func getRealIP(r *http.Request) string {
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		logrus.Debugf("resolve real ip from x-real-ip: %s", realIP)
		return realIP
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	logrus.Debugf("resolve real ip from remote addr: %s", tcpAddr.IP.String())
	return tcpAddr.IP.String()
}

func (m *WebsocketManager) Run() error {
	if (<-m.depOkSig) != 1 {
		return errors.New("dependency canal start failed")
	}
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
			m.responseUnauthenticated(ws, "sid not found")
			logrus.Warn("authentication failed: sid not found")
			return
		}

		res, err := m.hu60Client.GetProfile(context.Background(), hu60.GetProfileRequest{
			CommonRequest: hu60.CommonRequest{XFFIP: getRealIP(r)},
			Sid:           sid,
		})

		if err != nil {
			m.responseUnauthenticated(ws, err.Error())
			logrus.Warn("authentication failed: ", err.Error())
			return
		}

		err = ws.WriteMessage(websocket.TextMessage, []byte(`{"event": "connected"}`))
		if err != nil {
			logrus.Error(err)
			return
		}
		m.connMapUpdateLock.Lock()
		if _, ok := m.connMap[res.Uid]; !ok {
			m.connMap[res.Uid] = []*websocket.Conn{}
		}
		userConnCount := len(m.connMap[res.Uid])
		if userConnCount >= m.options.ConnectionLimitPerUser {
			m.closeConn(m.connMap[res.Uid][userConnCount-1], "reached the connections limit")
			m.connMap[res.Uid] = m.connMap[res.Uid][:userConnCount-1]
			logrus.Infof("user %s reached the connections limit, closed the first one", res.Name)
		}
		m.connMap[res.Uid] = append(m.connMap[res.Uid], ws)
		m.connMapUpdateLock.Unlock()
		if res.Uid < 0 {
			m.broadcastEventChan <- BotEvent{Event: "online", Data: UserOnlineStatus{
				UID:   res.Uid,
				Count: m.validConnCount(res.Uid),
			}}
		}
		logrus.Infof("user %s is connected, there are currently %d connections",
			res.Name, m.validConnCount(res.Uid))
		go m.connMessageListener(res, ws)
	})
	m.startConnEventSendTask()
	m.startEventBroadcastTask()
	m.startUserEventSendTask()
	logrus.Info("bot listening on ", m.options.Listen, " for interact now. websocket endpoint is /v1/ws")
	return http.ListenAndServe(m.options.Listen, nil)
}

func (m *WebsocketManager) closeConn(conn *websocket.Conn, msg string) {
	go func(conn *websocket.Conn) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Trace("close: ws conn already closed")
			}
		}()
		conn.WriteJSON(BotEvent{Event: "disconnecting", Data: msg})
		conn.Close()
	}(conn)
}

func (m *WebsocketManager) broadcast(event BotEvent) {
	for k, v := range m.connMap {
		if !m.eventSubscribed(k, event.Event) {
			continue
		}
		for _, ws := range v {
			go func(ws *websocket.Conn) {
				defer func() {
					if err := recover(); err != nil {
						logrus.Trace("ws conn already closed")
					}
				}()
				ws.WriteJSON(event)
				// ignore error, we only broadcast to valid connections
			}(ws)
		}
	}
}

func (m *WebsocketManager) startEventBroadcastTask() {
	go func() {
		defer close(m.broadcastEventChan)
		for e := range m.broadcastEventChan {
			m.broadcast(e)
		}
	}()
}

func (m *WebsocketManager) startConnEventSendTask() {
	go func() {
		defer close(m.connEventChan)
		for e := range m.connEventChan {
			if !m.eventSubscribed(e.Uid, e.Event.Event) {
				continue
			}
			go func(event ConnBotEvent) {
				defer func() {
					if err := recover(); err != nil {
						logrus.Trace("send conn event: ws conn already closed")
					}
				}()
				event.Ws.WriteJSON(event.Event)
			}(e)
		}
	}()
}

func (m *WebsocketManager) startUserEventSendTask() {
	go func() {
		defer close(m.userEventChan)
		for e := range m.userEventChan {
			if !m.eventSubscribed(e.Uid, e.Event.Event) {
				continue
			}
			if wsArr, ok := m.connMap[e.Uid]; ok {
				for _, conn := range wsArr {
					if conn == nil {
						continue
					}
					go func(ws *websocket.Conn, event BotEvent) {
						defer func() {
							if err := recover(); err != nil {
								logrus.Trace("send user event: ws conn already closed")
							}
						}()
						err := ws.WriteJSON(event)
						if err != nil {
							logrus.Warn("send user event error: ", err)
						}
					}(conn, e.Event)
				}
			} else {
				logrus.Infof("uid %d not online, discard event: %s", e.Uid, e.Event.Event)
			}
		}
	}()
}

func (m *WebsocketManager) OnCanalStartSucceed() {
	m.depOkSig <- 1
}

func (m *WebsocketManager) OnCanalStartFailed() {
	m.depOkSig <- 0
}

func (m *WebsocketManager) eventSubscribed(uid int, event string) bool {
	if unsubs, ok := m.unsubscribedEvents[uid]; ok {
		return !slices.Contains(unsubs, event)
	}
	return true
}

func (m *WebsocketManager) validConnCount(uid int) int {
	validConnCount := len(m.connMap[uid])
	for _, _ws := range m.connMap[uid] {
		if _ws == nil {
			validConnCount--
		}
	}
	return validConnCount
}

func (m *WebsocketManager) connMessageListener(userProfile hu60.GetProfileResponse, ws *websocket.Conn) {
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			logrus.Debugf("user %s, read message error: %w", userProfile.Name, err)
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				ws.Close()
			}
			m.connMapUpdateLock.Lock()
			validConnCount := len(m.connMap[userProfile.Uid])
			for i, _ws := range m.connMap[userProfile.Uid] {
				if _ws == nil {
					validConnCount--
				}
				if _ws == ws {
					m.connMap[userProfile.Uid][i] = nil
					validConnCount--
				}
			}
			if validConnCount == 0 {
				delete(m.connMap, userProfile.Uid)
			}
			m.connMapUpdateLock.Unlock()
			if userProfile.Uid < 0 {
				m.broadcastEventChan <- BotEvent{Event: "offline", Data: UserOnlineStatus{
					UID:   userProfile.Uid,
					Count: validConnCount,
				}}
			}
			logrus.Infof("user %s is disconnected, there are currently %d connections",
				userProfile.Name, validConnCount)
			break
		}
		m.processBotAction(ws, msg, userProfile.Uid)
	}
}

func (m *WebsocketManager) responseUnauthenticated(ws *websocket.Conn, msg string) {
	ws.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, msg))
	ws.Close()
}

func (m *WebsocketManager) responseError(ws *websocket.Conn, err error) {
	e := ws.WriteJSON(BotEvent{Event: "error", Data: err.Error()})
	if e != nil {
		logrus.Error(e.Error())
	}
}

func (m *WebsocketManager) processBotAction(ws *websocket.Conn, msg []byte, uid int) {
	var cmd BotCmd
	err := json.NewDecoder(strings.NewReader(string(msg))).Decode(&cmd)
	if err != nil {
		m.responseError(ws, err)
		return
	}
	if cmd.ID != "" {
		ws.WriteJSON(BotEvent{Event: "ack", Data: cmd.ID})
	}
	if action, ok := actions[cmd.Action]; ok {
		if slices.Contains(m.options.DisabledActions, cmd.Action) {
			m.responseError(ws, fmt.Errorf("disabled action: %s", cmd.Action))
			return
		}
		action(m, ws, cmd, uid)
		return
	}
	m.responseError(ws, fmt.Errorf("unsupported action: %s", cmd.Action))
}
