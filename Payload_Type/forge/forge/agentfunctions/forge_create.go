package agentfunctions

import (
	"encoding/json"
	"errors"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
	"os"
)

const assemblyGroup = "Create New .NET Assembly Command"
const bofGroup = "Create New BOF Command"

func init() {
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(agentstructs.Command{
		Name:                fmt.Sprintf("%s_create", PayloadTypeName),
		Description:         "Create brand new .NET or BOF commands to be available across all supported agent types.",
		HelpString:          fmt.Sprintf("%s_create", PayloadTypeName),
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{fmt.Sprintf("%s:create", PayloadTypeName)},
		ScriptOnlyCommand:   true,

		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:      []string{agentstructs.SUPPORTED_OS_WINDOWS},
			CommandIsBuiltin: true,
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "collectionName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE_CUSTOM,
				Description:      "Choose which collection to associate this new command with",
				ModalDisplayName: "Collection Name",
				DynamicQueryFunction: func(message agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
					return getCollectionSourceNameOptions(message)
				},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           assemblyGroup,
						UIModalPosition:     1,
					},
					{
						ParameterIsRequired: true,
						GroupName:           bofGroup,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "commandName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Specify which command to create",
				ModalDisplayName: "Command Name to Register",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           assemblyGroup,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:             "description",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Description of the new command",
				ModalDisplayName: "Command Description",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						GroupName:           assemblyGroup,
						UIModalPosition:     3,
					},
					{
						ParameterIsRequired: false,
						GroupName:           bofGroup,
						UIModalPosition:     3,
					},
				},
			},
			{
				Name:             "commandFileAssembly",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Upload the .net to execute for this command version",
				ModalDisplayName: "The assembly file to execute",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           assemblyGroup,
						UIModalPosition:     4,
					},
				},
			},
			{
				Name:             "commandVersion",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE_CUSTOM,
				Description:      "What version is this assembly",
				ModalDisplayName: "Version",
				DefaultValue:     "4.7_Any",
				DynamicQueryFunction: func(message agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
					return assemblyVersions
				},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						GroupName:           assemblyGroup,
						UIModalPosition:     5,
					},
				},
			},
			{
				Name:             "commandFilesBof",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE_MULTIPLE,
				Description:      "Upload the .o files to execute for this command",
				ModalDisplayName: "The bof .o files to execute",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           bofGroup,
						UIModalPosition:     4,
					},
				},
			},
			{
				Name:             "extensionFile",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "The extension.json file that describes this bof command",
				ModalDisplayName: "Sliver Armory style extension.json file",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           bofGroup,
						UIModalPosition:     5,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			var commandName string
			description, err := taskData.Args.GetStringArg("description")
			if err != nil {
				logging.LogError(err, "failed to get commandName")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			parameterGroup, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				logging.LogError(err, "failed to get parameterGroup")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if parameterGroup == assemblyGroup {
				commandName, err = taskData.Args.GetStringArg("commandName")
				if err != nil {
					logging.LogError(err, "failed to get commandName")
					response.Success = false
					response.Error = err.Error()
					return response
				}
			} else {
				// fetch the command name from the extension.json file that was uploaded
				extensionFileID, err := taskData.Args.GetFileArg("extensionFile")
				if err != nil {
					logging.LogError(err, "failed to get version")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				if extensionFileID == "" {
					response.Error = "No extension file specified"
					response.Success = false
					return response
				}
				contentResp, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
					AgentFileID: extensionFileID,
				})
				if err != nil {
					logging.LogError(err, "failed to send file content request to Mythic")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				if !contentResp.Success {
					response.Success = false
					response.Error = contentResp.Error
					return response
				}
				bofCommandExtension := bofCommandDefinition{}
				err = json.Unmarshal(contentResp.Content, &bofCommandExtension)
				if err != nil {
					logging.LogError(err, "failed to unmarshal extension.json file into struct")
					response.Success = false
					response.Error = contentResp.Error
					return response
				}
				commandName = bofCommandExtension.CommandName
			}
			collection, err := taskData.Args.GetStringArg("collectionName")
			if err != nil {
				logging.LogError(err, "failed to get collection name")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			collectionSourceData, err := getCollectionSource(collection)
			if errors.Is(err, collectionSourceNotFoundError) {
				if parameterGroup == assemblyGroup {
					collectionSourceData.Type = "assembly"
				} else {
					collectionSourceData.Type = "bof"
				}
				err = addCollectionSource(collectionSourceData)
				if err != nil {
					logging.LogError(err, "failed to add collection source")
					response.Success = false
					response.Error = err.Error()
					return response
				}
			} else if err != nil {
				logging.LogError(err, "failed to get collection source by name")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			fileContents, err := getOrCreateFile(collectionSourceData.SourceFilename)
			if err != nil {
				logging.LogError(err, "failed to get contents of collection sources file")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			commandSources := []collectionSourceCommandData{}
			err = json.Unmarshal(fileContents, &commandSources)
			if err != nil {
				logging.LogError(err, "failed to unmarshal contents of collection source file")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			displayParams := fmt.Sprintf("-collectionName %s -commandName %s", collection, commandName)
			response.DisplayParams = &displayParams
			if (collectionSourceData.Type == "assembly" && parameterGroup == bofGroup) ||
				(collectionSourceData.Type == "bof" && parameterGroup == assemblyGroup) {
				response.Success = false
				response.Error = fmt.Sprintf("This collection is of type %s, but you're trying to create a command of the wrong type.\nCreate a new collection or create a new command of the right type", collectionSourceData.Type)
				return response
			}
			commandIndex := -1
			newCommandSource := collectionSourceCommandData{
				Name:        commandName,
				CommandName: commandName,
				Description: description,
			}
			for i, commandSource := range commandSources {
				if commandSource.CommandName == commandName {
					if commandSource.RepoURL != "" || commandSource.CustomDownloadURL != "" {
						// trying to upload commandX when one already exists with a URL, not good
						response.Success = false
						response.Error = "Can't create new commandX when one already exists that references a remote URL"
						return response
					}
					// we already have this command name, but there's no remote url, so it was created like this
					// this is ok to update
					commandIndex = i
					newCommandSource = commandSource
					newCommandSource.Description = description
				}
			}
			var prefixedCommandName string
			if parameterGroup == assemblyGroup {
				commandFileID, err := taskData.Args.GetFileArg("commandFileAssembly")
				if err != nil {
					logging.LogError(err, "failed to get commandFile")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				commandVersion, err := taskData.Args.GetStringArg("commandVersion")
				if err != nil {
					logging.LogError(err, "failed to get version")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				prefixedCommandName = fmt.Sprintf("%s%s", AssemblyPrefix, commandName)
				newCommandSource.customAssemblyFileID = commandFileID
				newCommandSource.CustomVersion = commandVersion
				err = downloadAssemblyFile(newCommandSource, commandVersion, collectionSourceData, taskData)
				if err != nil {
					logging.LogError(err, "failed to download file to container")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("Registering new command %s\n", prefixedCommandName)),
				})
				newCommand := createAssemblyCommand(newCommandSource, collectionSourceData, true)
				agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(newCommand)
			} else {
				commandFileIDs, err := taskData.Args.GetArrayArg("commandFilesBof")
				if err != nil {
					logging.LogError(err, "failed to get commandFile")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				extensionFileID, err := taskData.Args.GetFileArg("extensionFile")
				if err != nil {
					logging.LogError(err, "failed to get version")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				if len(commandFileIDs) == 0 {
					response.Error = "No command files specified"
					response.Success = false
					return response
				}
				if extensionFileID == "" {
					response.Error = "No extension file specified"
					response.Success = false
					return response
				}
				prefixedCommandName = fmt.Sprintf("%s%s", BofPrefix, commandName)
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("Registering new command %s\n", prefixedCommandName)),
				})
				newCommandSource.customBofExtensionFileID = extensionFileID
				newCommandSource.customBofFileIDs = commandFileIDs
				err = downloadBofFile(newCommandSource, collectionSourceData, taskData)
				if err != nil {
					logging.LogError(err, "failed to download files to container")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				newCommand, err := createBofCommand(newCommandSource, collectionSourceData, true)
				if err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				}
				agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(newCommand)
			}
			rabbitmq.SyncPayloadData(&payloadDefinition.Name, false)
			response.Success = true
			// add / update command in sources file
			if commandIndex == -1 {
				commandSources = append(commandSources, newCommandSource)
			} else {
				commandSources[commandIndex] = newCommandSource
			}
			commandBytes, err := json.MarshalIndent(commandSources, "", "\t")
			if err != nil {
				logging.LogError(err, "failed to marshal command sources")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			err = os.WriteFile(collectionSourceData.SourceFilename, commandBytes, os.ModePerm)
			if err != nil {
				logging.LogError(err, "failed to marshal command sources")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: []byte(fmt.Sprintf("Command Registered for use!\n")),
			})
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			if len(input) > 0 {
				return args.LoadArgsFromJSONString(input)
			}
			return nil
		},
	})
}
