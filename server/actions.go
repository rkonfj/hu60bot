package server

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
)

var (
	actions map[string]func(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) = make(
		map[string]func(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int))

	startTime time.Time
)

var buildinActions []string

func init() {
	startTime = time.Now()

	actions["ls"] = listActionsAction
	actions["ping"] = pingAction
	actions["lsol"] = listOnlineUserAction
	actions["procinf"] = processInfoAction
	actions["unsub"] = unsubscribeEventAction
	actions["sub"] = subscribeEventAction
	actions["lsunsub"] = listUnsubscribedEventsAction

	actions["regapi"] = registerRobotApiAction
	actions["lsapi"] = listRobotApiAction
	actions["resp"] = robotApiRespAction

	for k := range actions {
		buildinActions = append(buildinActions, k)
	}
}

type ProcessInfo struct {
	StartedDuration string `json:"started"`
}

type RobotApiData struct {
	Uid    int    `json:"uid"`
	Action string `json:"action"`
	Data   any    `json:"data"`
}

type RobotApiData1 struct {
	Bot  int `json:"bot"`
	Data any `json:"data"`
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

func registerRobotApiAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	wm.robotActionHubLock.Lock()
	defer wm.robotActionHubLock.Unlock()
	if apis, ok := cmd.Data.([]any); ok {
		if _, ok := wm.robotActionHub[uid]; !ok {
			wm.robotActionHub[uid] = []string{}
		}
		for _, api := range apis {
			if apiStr, ok := api.(string); ok {
				if slices.Contains(wm.robotActionHub[uid], apiStr) {
					continue
				}
				if slices.Contains(buildinActions, apiStr) {
					ws.WriteJSON(BotEvent{Event: "error", Data: fmt.Sprintf("%s already exists. skiped", apiStr)})
					continue
				}
				wm.robotActionHub[uid] = append(wm.robotActionHub[uid], apiStr)
				actions[apiStr] = robotApiGenericAction
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

func listRobotApiAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	ws.WriteJSON(BotEvent{Event: cmd.Action, Data: wm.robotActionHub})
}

func robotApiGenericAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	if data, ok := cmd.Data.(map[string]any); ok {
		var (
			botAny any
			botUid float64
		)
		if botAny, ok = data["bot"]; !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "data.bot not found"})
			return
		}
		if botUid, ok = botAny.(float64); !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data.bot format"})
			return
		}
		req := BotEvent{Event: "req", Data: RobotApiData{Uid: uid, Action: cmd.Action, Data: data["data"]}}
		wm.userEventChan <- UserBotEvent{Uid: int(botUid), Event: req}
		return
	}
	ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data format"})
}

func robotApiRespAction(wm *WebsocketManager, ws *websocket.Conn, cmd BotCmd, uid int) {
	if data, ok := cmd.Data.(map[string]any); ok {
		var (
			clientUidAny any
			actionAny    any
			botDataAny   any
			clientUid    float64
			action       string
		)
		if clientUidAny, ok = data["uid"]; !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "data.uid not found"})
			return
		}
		if actionAny, ok = data["action"]; !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "data.action not found"})
			return
		}
		if botDataAny, ok = data["data"]; !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "data.data not found"})
			return
		}

		if clientUid, ok = clientUidAny.(float64); !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data.uid format"})
			return
		}

		if action, ok = actionAny.(string); !ok {
			ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data.action format"})
			return
		}
		resp := BotEvent{Event: action, Data: RobotApiData1{Bot: int(uid), Data: botDataAny}}
		wm.userEventChan <- UserBotEvent{Uid: int(clientUid), Event: resp}
		return
	}
	ws.WriteJSON(BotEvent{Event: "error", Data: "invalid data format"})
}
