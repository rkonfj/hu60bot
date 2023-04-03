package server

import (
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
)

var (
	actions map[string]func(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) = make(
		map[string]func(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int))

	startTime time.Time
)

func init() {
	startTime = time.Now()

	actions["ls"] = listActionsAction
	actions["ping"] = pingAction
	actions["lsol"] = listOnlineUserAction
	actions["procinf"] = processInfoAction
	actions["unsub"] = unsubscribeEventAction
	actions["sub"] = subscribeEventAction
	actions["lsunsub"] = listUnsubscribedEventsAction

}

type ProcessInfo struct {
	StartedDuration string `json:"started"`
}

type ActionDescription struct {
	Name string
	Desc string
}

func listActionsAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	var actionNames []string
	for k := range actions {
		if slices.Contains(wm.options.DisabledActions, k) {
			continue
		}
		actionNames = append(actionNames, k)
	}
	ws.WriteJSON(BotEvent{Event: cmd.Action, Data: actionNames})
}

func pingAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	ws.WriteJSON(BotEvent{Event: cmd.Action, Data: "pong"})
}

// 列出在线用户
func listOnlineUserAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	onlineData := make(map[int]int)
	for k, v := range wm.connMap {
		if k > 0 {
			continue
		}
		onlineData[k] = len(v)
	}
	ws.WriteJSON(BotEvent{Event: cmd.Action, Data: onlineData})
}

func processInfoAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	ws.WriteJSON(BotEvent{Event: cmd.Action, Data: ProcessInfo{
		StartedDuration: time.Since(startTime).String(),
	}})
}

func unsubscribeEventAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	wm.unsubscribedEventsLock.Lock()
	defer wm.unsubscribedEventsLock.Unlock()
	if opEvents, ok := cmd.Data.([]any); ok {
		if _, ok := wm.unsubscribedEvents[uid]; !ok {
			wm.unsubscribedEvents[uid] = []string{}
		}
		for _, e := range opEvents {
			if eStr, ok := e.(string); ok {
				if slices.Contains(wm.unsubscribedEvents[uid], eStr) {
					continue
				}
				wm.unsubscribedEvents[uid] = append(wm.unsubscribedEvents[uid], eStr)
				continue
			}
			ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data format"})
			return
		}
		ws.WriteJSON(BotEvent{Event: cmd.Action, Data: "ok"})
		return
	}
	ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data format"})
}

func subscribeEventAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	wm.unsubscribedEventsLock.Lock()
	defer wm.unsubscribedEventsLock.Unlock()
	if _, ok := wm.unsubscribedEvents[uid]; !ok {
		return
	}
	if opEvents, ok := cmd.Data.([]any); ok {
		var opEventStrs []string
		for _, e := range opEvents {
			if eStr, ok := e.(string); ok {
				opEventStrs = append(opEventStrs, eStr)
				continue
			}
			ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data format"})
			return
		}
		var events []string
		for _, e := range wm.unsubscribedEvents[uid] {
			if !slices.Contains(opEventStrs, e) {
				events = append(events, e)
			}
		}
		wm.unsubscribedEvents[uid] = events
		ws.WriteJSON(BotEvent{Event: cmd.Action, Data: "ok"})
		return
	}
	ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data format"})
}

func listUnsubscribedEventsAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	ws.WriteJSON(BotEvent{Event: cmd.Action, Data: wm.unsubscribedEvents[uid]})
}
