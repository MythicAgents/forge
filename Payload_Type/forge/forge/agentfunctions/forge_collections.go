package agentfunctions

import (
	"encoding/json"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"os"
	"path/filepath"
)

func init() {
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(agentstructs.Command{
		Name:                fmt.Sprintf("%s_collections", PayloadTypeName),
		Description:         fmt.Sprintf("Interact with the %s container to view available commands.", PayloadTypeName),
		HelpString:          fmt.Sprintf("%s_collections", PayloadTypeName),
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		ScriptOnlyCommand:   true,
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", PayloadTypeName, "browserscripts", fmt.Sprintf("%s_collections.js", PayloadTypeName)),
			Author:     "@its_a_feature_",
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:      []string{agentstructs.SUPPORTED_OS_WINDOWS},
			CommandIsBuiltin: true,
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "collectionName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE_CUSTOM,
				Description:      "Choose which collection to query",
				ModalDisplayName: "Collection Name to Query",
				DynamicQueryFunction: func(message agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
					return getCollectionSourceNameOptions(message)
				},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			collection, err := taskData.Args.GetChooseOneArg("collectionName")
			if err != nil {
				logging.LogError(err, "failed to get collection name")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			displayParams := fmt.Sprintf("-collectionName %s", collection)
			response.DisplayParams = &displayParams

			collectionSourceData, err := getCollectionSource(collection)
			if err != nil {
				logging.LogError(err, "failed to get collection source")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			fileContents, err := getOrCreateFile(collectionSourceData.SourceFilename)
			if err != nil {
				logging.LogError(err, "failed to get collection source file contents")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			commandSources := []collectionSourceCommandData{}
			err = json.Unmarshal(fileContents, &commandSources)
			if err != nil {
				logging.LogError(err, "failed to unmarshal collection source file contents")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			commandNames := make([]string, len(commandSources))
			for i, _ := range commandSources {
				switch collectionSourceData.Type {
				case "assembly":
					commandNames[i] = fmt.Sprintf("%s%s", AssemblyPrefix, commandSources[i].CommandName)
				case "bof":
					commandNames[i] = fmt.Sprintf("%s%s", BofPrefix, commandSources[i].CommandName)
				}

			}
			commandSearchResp, err := mythicrpc.SendMythicRPCCallbackSearchCommand(mythicrpc.MythicRPCCallbackSearchCommandMessage{
				CallbackID:         &taskData.Callback.ID,
				SearchCommandNames: &commandNames,
			})
			if err != nil {
				logging.LogError(err, "failed to send search for commands commands")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if !commandSearchResp.Success {
				logging.LogError(err, "got error message back from command search from Mythic")
				response.Success = false
				response.Error = commandSearchResp.Error
				return response
			}
			for i, _ := range commandSources {
				commandSources[i].CollectionName = collection
				if commandSources[i].RepoURL != "" || commandSources[i].CustomDownloadURL != "" {
					commandSources[i].Downloadable = true
				}
				switch collectionSourceData.Type {
				case "assembly":
					commandSources[i].CommandName = fmt.Sprintf("%s%s", AssemblyPrefix, commandSources[i].CommandName)
					oneExists := false
					if commandSources[i].CustomVersion != "" {
						commandFilePath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSources[i].CustomVersion, commandSources[i].Name+".exe")
						_, err = os.Stat(commandFilePath)
						if err == nil {
							oneExists = true
						}
					} else {
						for _, ver := range assemblyVersions {
							commandFilePath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, ver, commandSources[i].Name+".exe")
							_, err = os.Stat(commandFilePath)
							if err == nil {
								oneExists = true
								break
							}
						}
					}
					if oneExists {
						commandSources[i].Downloaded = true
					}

				case "bof":
					commandSources[i].CommandName = fmt.Sprintf("%s%s", BofPrefix, commandSources[i].CommandName)
					_, err = os.Stat(filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSources[i].Name, "extension.json"))
					if err == nil {
						commandSources[i].Downloaded = true
					}
				}
				for _, registeredCommand := range commandSearchResp.Commands {
					switch collectionSourceData.Type {
					case "assembly":
						if commandSources[i].CommandName == registeredCommand.Name {
							commandSources[i].Registered = true
							break
						}
					case "bof":
						if commandSources[i].CommandName == registeredCommand.Name {
							commandSources[i].Registered = true
							break
						}
					}

				}
			}
			commandOptionBytes, err := json.Marshal(&commandSources)
			if err != nil {
				logging.LogError(err, "failed to marshal updated command sources back to bytes")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			resp, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: commandOptionBytes,
			})
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if !resp.Success {
				response.Success = false
				response.Error = resp.Error
				return response
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
	})
}
