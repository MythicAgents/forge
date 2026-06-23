package agentfunctions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
)

func removeCommandFromFile(commandSource collectionSourceCommandData, collectionSourceData collectionSource) error {
	assemblyCommandsFile, err := getOrCreateFile(collectionSourceData.CommandsFilename)
	if err != nil {
		logging.LogError(err, "Failed to read assembly commands file")
		return err
	}
	if collectionSourceData.Type == "assembly" {
		commands := []assemblyCommand{}
		err = json.Unmarshal(assemblyCommandsFile, &commands)
		if err != nil {
			logging.LogError(err, "failed to parse assembly commands into struct")
			return err
		}
		for i, _ := range commands {
			if commands[i].CommandName == fmt.Sprintf("%s%s", AssemblyPrefix, commandSource.CommandName) {
				// we found the one to remove
				commands = append(commands[:i], commands[i+1:]...)
				newAssemblyCommandsBytes, err := json.MarshalIndent(commands, "", "\t")
				if err != nil {
					logging.LogError(err, "failed to marshal commands into JSON")
					return err
				}
				err = os.WriteFile(collectionSourceData.CommandsFilename, newAssemblyCommandsBytes, os.ModePerm)
				if err != nil {
					logging.LogError(err, "failed to write out new commands to file")
					return err
				}
				return nil
			}
		}
		// never found the command, so it's essentially removed
		return nil
	} else if collectionSourceData.Type == "bof" {
		commands := []bofCommand{}
		err = json.Unmarshal(assemblyCommandsFile, &commands)
		if err != nil {
			logging.LogError(err, "failed to parse assembly commands into struct")
			return err
		}
		commandNamesToRemove := make(map[string]bool)
		for _, commandName := range getBofCommandNamesForRemoval(commandSource, collectionSourceData) {
			commandNamesToRemove[commandName] = true
		}
		filteredCommands := make([]bofCommand, 0, len(commands))
		removedCommand := false
		for _, command := range commands {
			if commandNamesToRemove[command.CommandName] || command.CollectionCommandName == commandSource.Name {
				removedCommand = true
				continue
			}
			filteredCommands = append(filteredCommands, command)
		}
		if !removedCommand {
			// never found the command, so it's essentially removed
			return nil
		}
		newAssemblyCommandsBytes, err := json.MarshalIndent(filteredCommands, "", "\t")
		if err != nil {
			logging.LogError(err, "failed to marshal commands into JSON")
			return err
		}
		err = os.WriteFile(collectionSourceData.CommandsFilename, newAssemblyCommandsBytes, os.ModePerm)
		if err != nil {
			logging.LogError(err, "failed to write out new commands to file")
			return err
		}
		return nil

	}
	return errors.New("unknown source type")
}

func init() {
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(agentstructs.Command{
		Name:                fmt.Sprintf("%s_register", PayloadTypeName),
		Description:         "Register existing possible commands to be available across all supported agent types.",
		HelpString:          fmt.Sprintf("%s_register -collectionName SharpCollection -commandName Rubeus", PayloadTypeName),
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{fmt.Sprintf("%s:register", PayloadTypeName)},
		ScriptOnlyCommand:   true,

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
			{
				Name:             "commandName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Specify which command to register",
				ModalDisplayName: "Command Name to Register",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
					},
				},
			},
			{
				Name:             "remove",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Unregister the command across all callbacks",
				ModalDisplayName: "Unregister the command across all callbacks",
				DefaultValue:     false,
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
			commandName, err := taskData.Args.GetStringArg("commandName")
			if err != nil {
				logging.LogError(err, "failed to get commandName")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			remove, err := taskData.Args.GetBooleanArg("remove")
			if err != nil {
				logging.LogError(err, "failed to get commandName")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			displayParams := fmt.Sprintf("-collectionName %s -commandName %s", collection, commandName)
			if remove {
				displayParams += " -remove"
			}
			response.DisplayParams = &displayParams
			collectionSourceData, err := getCollectionSource(collection)
			if err != nil {
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
			prefixedCommandNames := []string{}
			for _, commandSource := range commandSources {
				if commandSource.Name == commandName {
					switch collectionSourceData.Type {
					case "assembly":
						prefixedCommandName := fmt.Sprintf("%s%s", AssemblyPrefix, commandSource.CommandName)
						prefixedCommandNames = []string{prefixedCommandName}
						if remove {
							mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
								TaskID:   taskData.Task.ID,
								Response: []byte(fmt.Sprintf("Removing command %s\n", prefixedCommandName)),
							})
							err = removeCommandFromFile(commandSource, collectionSourceData)
							if err != nil {
								logging.LogError(err, "failed to remove command")
								response.Success = false
								response.Error = err.Error()
							}
							agentstructs.AllPayloadData.Get(PayloadTypeName).RemoveCommand(agentstructs.Command{Name: prefixedCommandName})
						} else {
							mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
								TaskID:   taskData.Task.ID,
								Response: []byte(fmt.Sprintf("Registering new command %s\n", prefixedCommandName)),
							})
							newCommand := createAssemblyCommand(commandSource, collectionSourceData, true)
							addOrReplaceForgeCommand(newCommand)
						}

					case "bof":
						prefixedCommandNames = getBofCommandNamesForSource(commandSource, collectionSourceData)
						prefixedCommandNamesText := strings.Join(prefixedCommandNames, ", ")
						if remove {
							prefixedCommandNames = getBofCommandNamesForRemoval(commandSource, collectionSourceData)
							prefixedCommandNamesText = strings.Join(prefixedCommandNames, ", ")
							mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
								TaskID:   taskData.Task.ID,
								Response: []byte(fmt.Sprintf("Removing command(s) %s\n", prefixedCommandNamesText)),
							})
							err = removeCommandFromFile(commandSource, collectionSourceData)
							if err != nil {
								logging.LogError(err, "failed to remove command")
								response.Success = false
								response.Error = err.Error()
							}
							for _, prefixedCommandName := range prefixedCommandNames {
								agentstructs.AllPayloadData.Get(PayloadTypeName).RemoveCommand(agentstructs.Command{Name: prefixedCommandName})
							}
						} else {
							mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
								TaskID:   taskData.Task.ID,
								Response: []byte(fmt.Sprintf("Registering new command(s) %s\n", prefixedCommandNamesText)),
							})
							err = createBofCommand(commandSource, collectionSourceData, true)
							if err != nil {
								response.Success = false
								response.Error = err.Error()
								return response
							}
						}

					default:
					}
					rabbitmq.SyncPayloadData(&payloadDefinition.Name, false)
					response.Success = true
					if remove {

						// now to remove this command from all associated callbacks
						payloadtypesFileContents, err := getOrCreateFile(PayloadTypeSupportFilename)
						if err != nil {
							logging.LogError(err, "failed to read payloadtype support file")
							response.Success = false
							response.Error = err.Error()
							return response
						}
						payloadTypes := []agentDefinition{}
						err = json.Unmarshal(payloadtypesFileContents, &payloadTypes)
						if err != nil {
							logging.LogError(err, "failed to read unmarshal payloadtypes file")
							response.Success = false
							response.Error = err.Error()
							return response
						}
						payloadTypeNames := make([]string, len(payloadTypes))
						for i, payloadType := range payloadTypes {
							payloadTypeNames[i] = payloadType.Agent
						}
						callbacksSearchResp, err := mythicrpc.SendMythicRPCCallbackSearch(mythicrpc.MythicRPCCallbackSearchMessage{
							AgentCallbackID:            taskData.Callback.AgentCallbackID,
							SearchCallbackPayloadTypes: &payloadTypeNames,
						})
						if err != nil {
							logging.LogError(err, "failed to send mythicrpc message to mythic to search for callbacks")
							response.Success = false
							response.Error = err.Error()
							return response
						}
						if !callbacksSearchResp.Success {
							logging.LogError(nil, "mythicrpc returned error", "error", callbacksSearchResp.Error)
							response.Success = false
							response.Error = callbacksSearchResp.Error
							return response
						}
						callbackIDs := make([]int, len(callbacksSearchResp.Results))
						for i, callback := range callbacksSearchResp.Results {
							callbackIDs[i] = callback.ID
						}
						callbacksRemoveCommandResp, err := mythicrpc.SendMythicRPCCallbackRemoveCommand(mythicrpc.MythicRPCCallbackRemoveCommandMessage{
							TaskID:      taskData.Task.ID,
							PayloadType: PayloadTypeName,
							CallbackIDs: callbackIDs,
							Commands:    prefixedCommandNames,
						})
						if err != nil {
							logging.LogError(err, "failed to send mythicrpc message to mythic to remove commands")
							response.Success = false
							response.Error = err.Error()
							return response
						}
						if !callbacksRemoveCommandResp.Success {
							logging.LogError(nil, "mythicrpc returned error", "error", callbacksSearchResp.Error)
							response.Success = false
							response.Error = callbacksSearchResp.Error
							return response
						}
						mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
							TaskID:   taskData.Task.ID,
							Response: []byte(fmt.Sprintf("Command Removed from use!\n")),
						})
					} else {
						mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
							TaskID:   taskData.Task.ID,
							Response: []byte(fmt.Sprintf("Command Registered for use!\n")),
						})
					}
					return response
				}
			}
			response.Success = false
			response.Error = "Failed to find that command in " + collectionSourceData.SourceFilename
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
