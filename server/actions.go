package server

import (
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	actions map[string]func(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) = make(
		map[string]func(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int))
)

func init() {
	actions["chat"] = chatAction
	actions["rmconvo"] = removeConvoAction
	actions["ping"] = pingAction
	actions["lsol"] = listOnlineUserAction
}

func chatAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	conversationKey := fmt.Sprintf("%d", uid)
	if d, ok := cmd.Data.(string); ok {
		responseText, newConversation, err := wm.cm.Ask(d, conversationKey)
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
}

// 清除 AI 会话
func removeConvoAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	conversationKey := fmt.Sprintf("%d", uid)
	wm.cm.MarkExpired(conversationKey)
	ws.WriteJSON(BotEvent{Event: "rmconvo", Data: "ok"})
}

func pingAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	ws.WriteJSON(BotEvent{Event: "ping", Data: "pong"})
}

// 列出在线用户
func listOnlineUserAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	onlineData := make(map[int]int)
	for k, v := range wm.connMap {
		onlineData[k] = len(v)
	}
	ws.WriteJSON(BotEvent{Event: "lsol", Data: onlineData})
}
