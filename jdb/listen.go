package jdb

import (
	"time"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
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

func Listen(url, channel, tag string, listen func(res et.Json)) {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			console.Error(err)
		}
	}

	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(url, minReconn, maxReconn, reportProblem)
	listenEvent(url, channel, tag, listener, listen)
}

func listenEvent(url, channel, tag string, listener *pq.Listener, listen func(res et.Json)) {
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
		console.Panic(err)
	}

	console.LogKF("Listen", "channel:%s", channel)

	for isCloseListen(channel) {
		notification(listener, channel, tag, listen)
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

	console.LogF("DB stop channel:%s", channel)
}

func isCloseListen(channel string) bool {
	key := strs.Format(`%s`, channel)
	result := closeListen == key
	return !result
}

func notification(l *pq.Listener, channel, tag string, listen func(res et.Json)) {
	select {
	case n := <-l.Notify:
		result, err := et.ToJson(n.Extra)
		if err != nil {
			console.AlertF("notification: Not conver to et.Json channel:%s result:%s", channel, n.Extra)
		}

		result.Set("tag", tag)
		listen(result)
	case <-time.After(90 * time.Second):
		go l.Ping()
	}
}
