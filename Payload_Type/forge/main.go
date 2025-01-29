package main

import (
	"community_collections/forge/agentfunctions"
	"github.com/MythicMeta/MythicContainer"
	"github.com/MythicMeta/MythicContainer/logging"
	"os"
)

func main() {
	agentfunctions.Initialize()
	if len(os.Args) > 1 {
		if os.Args[1] == "download" {
			logging.UpdateLogToStdout("debug")
			agentfunctions.DownloadEverything()
			return
		}
	}
	MythicContainer.StartAndRunForever([]MythicContainer.MythicServices{
		MythicContainer.MythicServicePayload,
	})
}
