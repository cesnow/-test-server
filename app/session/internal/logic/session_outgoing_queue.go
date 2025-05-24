package logic

import (
	"container/list"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"math"
)

const (
	maxAckedIdListSize = 500
)

type outboxMsg struct {
	msgId int64
	sent  int64
	state byte
	msg   *tproto.MsgRawData
}

type sessionOutgoingQueue struct {
	minMsgId    int64
	maxMsgId    int64
	listOutMsg  *list.List
	ackedIdList []int64
}

func newSessionOutgoingQueue() *sessionOutgoingQueue {
	return &sessionOutgoingQueue{
		minMsgId:    0,
		maxMsgId:    math.MaxInt64,
		listOutMsg:  list.New(),
		ackedIdList: make([]int64, 0),
	}
}

func (q *sessionOutgoingQueue) AddRpcResultMessage(reqMsgId int64, result *tproto.MsgRawData) *outboxMsg {
	oMsg := q.Lookup(reqMsgId)
	if oMsg == nil {
		oMsg = &outboxMsg{
			msgId: reqMsgId,
			sent:  0,
			state: ACKNOWLEDGED,
			msg:   result,
		}
		q.listOutMsg.PushBack(oMsg)
	}

	q.Shrink()
	return oMsg
}

func (q *sessionOutgoingQueue) AddNotifyMessage(notifyId int64, confirm bool, msg *tproto.MsgRawData) *outboxMsg {
	oMsg := new(outboxMsg)
	oMsg.msgId = notifyId
	oMsg.sent = 0
	if confirm {
		oMsg.state = ACKNOWLEDGED
	} else {
		oMsg.state = NO_NEED_ACK
	}
	oMsg.msg = msg
	q.listOutMsg.PushBack(oMsg)

	q.Shrink()
	return oMsg
}

func (q *sessionOutgoingQueue) AddPushUpdates(pushMsgId int64, result *tproto.MsgRawData) *outboxMsg {
	oMsg := q.Lookup(pushMsgId)
	if oMsg == nil {
		oMsg = &outboxMsg{
			msgId: pushMsgId,
			sent:  0,
			state: ACKNOWLEDGED,
			msg:   result,
		}
		q.listOutMsg.PushBack(oMsg)
	}

	q.Shrink()
	return oMsg
}

func (q *sessionOutgoingQueue) OnMessagesAck(ackIds []int64, cb func(inMsgId int64)) {
	var next *list.Element
	for _, id := range ackIds {
		for e := q.listOutMsg.Front(); e != nil; e = next {
			next = e.Next()
			if id == e.Value.(*outboxMsg).msg.MsgId {
				iMsgId := e.Value.(*outboxMsg).msgId
				q.ackedIdList = append(q.ackedIdList, iMsgId)
				cb(iMsgId)
				q.listOutMsg.Remove(e)
			}
		}
	}

	if len(q.ackedIdList) > maxAckedIdListSize {
		q.ackedIdList = q.ackedIdList[len(q.ackedIdList)-maxAckedIdListSize-1:]
	}
}

func (q *sessionOutgoingQueue) Lookup(msgId int64) (oMsg *outboxMsg) {
	for e := q.listOutMsg.Front(); e != nil; e = e.Next() {
		if msgId == e.Value.(*outboxMsg).msgId {
			oMsg = e.Value.(*outboxMsg)
		}
	}
	return
}

func (q *sessionOutgoingQueue) Remove(msgId int64) (oMsg *outboxMsg) {
	for e := q.listOutMsg.Front(); e != nil; e = e.Next() {
		if msgId == e.Value.(*outboxMsg).msgId {
			oMsg = e.Value.(*outboxMsg)
			q.listOutMsg.Remove(e)
			break
		}
	}
	return
}

func (q *sessionOutgoingQueue) Shrink() {
	for q.listOutMsg.Len() > maxQueueSize {
		iMsg := q.listOutMsg.Remove(q.listOutMsg.Front())
		q.minMsgId = iMsg.(*outboxMsg).msgId
	}
}
