package master

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
)

func listenSync(res et.Json) {
	idT := res.Str("_idt")
	nodeId := res.Str("nodo")

	node := master.GetNodeByID(nodeId)
	if node == nil {
		return
	}

	go node.SyncIdT(idT)
}

func listenNode(res et.Json) {
	action := res.Str("action")
	nodeId := res.Str("nodo")

	switch strs.Uppcase(action) {
	case "INSERT":
		go master.LoadNodeById(nodeId)
	case "UPDATE":
		go master.LoadNodeById(nodeId)
	case "DELETE":
		go master.UnloadNodeById(nodeId)
	}
}
