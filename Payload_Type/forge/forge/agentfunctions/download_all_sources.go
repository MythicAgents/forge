package agentfunctions

import (
	"encoding/json"
	"github.com/MythicMeta/MythicContainer/logging"
	"sync"
)

func DownloadEverything() {
	collections := getCollectionSources()
	wg := sync.WaitGroup{}
	for _, collectionSourceData := range collections {
		commandSources := []collectionSourceCommandData{}
		commandSourcesFileData, err := getOrCreateFile(collectionSourceData.SourceFilename)
		err = json.Unmarshal(commandSourcesFileData, &commandSources)
		if err != nil {
			logging.LogError(err, "failed to unmarshal contents of collection source file")
			continue
		}
		for _, commandSource := range commandSources {
			wg.Add(1)
			go func() {
				defer wg.Done()
				switch collectionSourceData.Type {
				case "assembly":
					if commandSource.CustomDownloadURL != "" {
						logging.LogInfo("[*] Starting download", "source", collectionSourceData.Name,
							"command", commandSource.Name, "version", commandSource.CustomVersion)
						err = downloadAssemblyFile(commandSource, commandSource.CustomVersion, collectionSourceData, nil)
						if err != nil {
							logging.LogError(err, "[!] failed to download assembly file", "source", collectionSourceData.Name,
								"command", commandSource.Name, "version", commandSource.CustomVersion)
						} else {
							logging.LogInfo("[*] Successfully downloaded", "source", collectionSourceData.Name, "command", commandSource.Name, "version", commandSource.CustomVersion)
						}
					} else {
						atLeastOneSuccess := false
						if commandSource.RepoURL == "" {
							logging.LogError(nil, "[!] No custom url and now repo url", "source", collectionSourceData.Name, "command", commandSource.Name)
							return
						}
						for _, assemblyVersion := range assemblyVersions {
							logging.LogInfo("[*] Starting download", "source", collectionSourceData.Name,
								"command", commandSource.Name, "version", assemblyVersion)
							err = downloadAssemblyFile(commandSource, assemblyVersion, collectionSourceData, nil)
							if err == nil {
								atLeastOneSuccess = true
								logging.LogInfo("[*] Successfully downloaded", "source", collectionSourceData.Name, "command", commandSource.Name, "version", assemblyVersion)
							} else {
								logging.LogError(err, "[!] failed to download assembly file", "source", collectionSourceData.Name,
									"command", commandSource.Name, "version", assemblyVersion)
							}
						}
						if !atLeastOneSuccess {
							logging.LogError(err, "[!] failed to download assembly file", "source", collectionSourceData.Name,
								"command", commandSource.Name)
						}
					}
				case "bof":
					logging.LogInfo("[*] Starting download", "source", collectionSourceData.Name,
						"command", commandSource.Name, "version", "bof")
					err = downloadBofFile(commandSource, collectionSourceData, nil)
					if err != nil {
						logging.LogError(err, "[!] failed to download bof file", "source", collectionSourceData.Name,
							"command", commandSource.Name)
					} else {
						logging.LogInfo("[*] Successfully downloaded", "source", collectionSourceData.Name, "command", commandSource.Name, "version", "bof")
					}
				}
			}()
		}
	}
	wg.Wait()
}
