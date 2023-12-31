package jdb

import (
	"time"

	"github.com/cgalvisleon/elvis/console"
	e "github.com/cgalvisleon/elvis/json"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/lib/pq"
)

var closeListen string = ""

func ListenClose(listen *pq.Listener) error {
	if listen == nil {
		return nil
	}

	err := listen.Close()
	if err != nil {
		return err
	}

	return nil
}

func Listen(nodo, url, channel string, listen func(res e.Json)) {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			console.Error(err)
		}
	}

	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(url, minReconn, maxReconn, reportProblem)
	ListenEvent(nodo, url, channel, listener, listen)
}

func ListenEvent(nodo, url, channel string, listener *pq.Listener, listen func(res e.Json)) {
	if url == "" {
		return
	}

	if channel == "" {
		return
	}

	if listen == nil {
		return
	}

	err := listener.Listen(channel)
	if err != nil {
		console.Panic("Listen", err)
	}

	console.LogKF("DB channel", "Start channel:%s nodo:%s", channel, nodo)

	for IsCloseListen(nodo, channel) {
		hostNotification(listener, channel, nodo, listen)
	}
	closeListen = ""

	err = listener.UnlistenAll()
	if err != nil {
		console.Error(err)
	}

	err = listener.Close()
	if err != nil {
		console.Error(err)
	}

	console.LogKF("DB channel", "Stop channel:%s nodo:%s", channel, nodo)
}

func CloseListen(host, channel string) {
	closeListen = strs.Format(`%s/%s`, host, channel)
}

func IsCloseListen(host, channel string) bool {
	key := strs.Format(`%s/%s`, host, channel)
	result := closeListen == key
	return !result
}

func hostNotification(l *pq.Listener, channel string, nodo string, listen func(res e.Json)) {
	select {
	case n := <-l.Notify:
		result, err := e.ToJson(n.Extra)
		if err != nil {
			console.LogC("DB channel", "Red", strs.Format("hostNotification: Not conver to e.Json nodo:%s channel:%s result:%s", nodo, channel, n.Extra))
		}

		result["nodo"] = nodo
		listen(result)
	case <-time.After(90 * time.Second):
		go l.Ping()
	}
}
