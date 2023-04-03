package server

import (
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
)

type CanalManager struct {
	wsm       *WebsocketManager
	connector *client.SimpleCanalConnector
}

func NewCanalManager(opts CanalOptions, wsm *WebsocketManager) *CanalManager {
	return &CanalManager{
		wsm:       wsm,
		connector: client.NewSimpleCanalConnector(opts.CanalHost, opts.CanalPort, "", "", opts.CanalClientDestination, 60000, 60*60*1000),
	}
}

func (m *CanalManager) Run() error {
	err := m.connector.Connect()
	if err != nil {
		return err
	}
	err = m.connector.Subscribe("hu60\\.hu60_msg")
	if err != nil {
		return err
	}
	logrus.Info("bot watching db event now")
	m.wsm.OnCanalStartSucceed()
	for {

		message, err := m.connector.Get(100, nil, nil)
		if err != nil {
			logrus.Fatalf("bot exited. canal disconnected(%s)", err)
		}

		batchId := message.Id
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		processHu60Msg(message.Entries, func(msg *Hu60Msg) {
			m.wsm.Push(msg)
		})
	}
}

func processHu60Msg(entries []pbe.Entry, msgHandler func(msg *Hu60Msg)) {
	for _, entry := range entries {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		if err != nil {
			logrus.Error(err)
		}
		eventType := rowChange.GetEventType()
		if eventType != pbe.EventType_INSERT {
			logrus.Debug("canalManager: discard non insert event: ", eventType)
			continue
		}
		header := entry.GetHeader()
		logrus.Debugf("binlog[%s : %d],name[%s,%s], eventType: %s", header.GetLogfileName(), header.GetLogfileOffset(), header.GetSchemaName(), header.GetTableName(), header.GetEventType())

		for _, rowData := range rowChange.GetRowDatas() {
			var msg Hu60Msg
			for _, col := range rowData.AfterColumns {
				switch col.GetName() {
				case "id":
					intVal, _ := strconv.Atoi(col.GetValue())
					msg.ID = intVal
				case "byuid":
					intVal, _ := strconv.Atoi(col.GetValue())
					msg.ByUID = intVal
				case "touid":
					intVal, _ := strconv.Atoi(col.GetValue())
					msg.ToUID = intVal
				case "type":
					intVal, _ := strconv.Atoi(col.GetValue())
					msg.Type = intVal
				case "read":
					intVal, _ := strconv.Atoi(col.GetValue())
					msg.Read = intVal
				case "content":
					msg.Content = col.GetValue()
				}
			}
			msgHandler(&msg)
		}
	}
}
