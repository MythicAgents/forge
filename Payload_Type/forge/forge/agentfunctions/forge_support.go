package agentfunctions

import (
	"encoding/json"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
	"os"
)

func init() {
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(agentstructs.Command{
		Name:                fmt.Sprintf("%s_support", PayloadTypeName),
		Description:         fmt.Sprintf("Add, remove, or update support for a payload type with %s", PayloadTypeName),
		HelpString:          fmt.Sprintf("%s_support", PayloadTypeName),
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		ScriptOnlyCommand:   true,
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:      []string{agentstructs.SUPPORTED_OS_WINDOWS},
			CommandIsBuiltin: true,
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "payload_type",
				CLIName:          "payloadType",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE_CUSTOM,
				Description:      "Choose which payload type to modify support for",
				ModalDisplayName: "Payload Type",
				DefaultValue:     "",
				DynamicQueryFunction: func(message agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
					supportedAgentsFile, err := getOrCreateFile(PayloadTypeSupportFilename)
					if err != nil {
						logging.LogError(err, "failed to get supported payload types file")
						return []string{}
					}
					supportedAgents := []agentDefinition{}
					err = json.Unmarshal(supportedAgentsFile, &supportedAgents)
					if err != nil {
						logging.LogError(err, "failed to unmarshal supported payload types file")
						return []string{}
					}
					supportedAgentNames := make([]string, len(supportedAgents))
					for i, agent := range supportedAgents {
						supportedAgentNames[i] = agent.Agent
					}
					return supportedAgentNames
				},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     0,
					},
				},
			},
			{
				Name:             "bof_command",
				CLIName:          "bofCommand",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the bof command for this agent",
				ModalDisplayName: "BOF Command",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "bof_file_parameter_name",
				CLIName:          "bofFileParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies the bof file UUID",
				ModalDisplayName: "BOF File Parameter Name",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:             "bof_argument_array_parameter_name",
				CLIName:          "bofArgumentArrayParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies array of BOF arguments",
				ModalDisplayName: "BOF Arguments Array Parameter Name",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
					},
				},
			},
			{
				Name:             "bof_entrypoint_parameter_name",
				CLIName:          "bofEntrypointParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies the bof entrypoint function name",
				ModalDisplayName: "BOF Entrypoint Parameter Name",
				DefaultValue:     "go",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
					},
				},
			},
			{
				Name:             "inline_assembly_command",
				CLIName:          "inlineAssemblyCommand",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the command that runs assemblies inline",
				ModalDisplayName: "Inline Assembly Command Name",
				DefaultValue:     "inline_assembly",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     5,
					},
				},
			},
			{
				Name:             "inline_assembly_file_parameter_name",
				CLIName:          "inlineAssemblyFileParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies the assembly file UUID",
				ModalDisplayName: "Inline Assembly File Parameter Name",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
					},
				},
			},
			{
				Name:             "inline_assembly_argument_parameter_name",
				CLIName:          "inlineAssemblyArgumentParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies the assembly argument string",
				ModalDisplayName: "Inline Assembly argument string Parameter Name",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     7,
					},
				},
			},
			{
				Name:             "execute_assembly_command",
				CLIName:          "executeAssemblyCommand",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the command that runs assemblies in fork-and-run",
				ModalDisplayName: "Execute Assembly Command Name",
				DefaultValue:     "execute_assembly",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     8,
					},
				},
			},
			{
				Name:             "execute_assembly_file_parameter_name",
				CLIName:          "executeAssemblyFileParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies the assembly file UUID",
				ModalDisplayName: "Execute Assembly File Parameter Name",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     9,
					},
				},
			},
			{
				Name:             "execute_assembly_argument_parameter_name",
				CLIName:          "executeAssemblyArgumentParameterName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the parameter that specifies the assembly argument string",
				ModalDisplayName: "Execute Assembly argument string Parameter Name",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     10,
					},
				},
			},
			{
				Name:             "remove_support",
				CLIName:          "remove",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Remove this agent from the supported list",
				ModalDisplayName: "Remove Support",
				DefaultValue:     false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     11,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			inputAgent, _ := taskData.Args.GetStringArg("payload_type")
			inputBofCommand, _ := taskData.Args.GetStringArg("bof_command")
			inputBofFileParameterName, _ := taskData.Args.GetStringArg("bof_file_parameter_name")
			inputBofArgumentArrayParameterName, _ := taskData.Args.GetStringArg("bof_argument_array_parameter_name")
			inputBofEntrypointParameterName, _ := taskData.Args.GetStringArg("bof_entrypoint_parameter_name")
			inputInlineAssemblyCommand, _ := taskData.Args.GetStringArg("inline_assembly_command")
			inputInlineAssemblyFileParameterName, _ := taskData.Args.GetStringArg("inline_assembly_file_parameter_name")
			inputInlineAssemblyArgumentParameterName, _ := taskData.Args.GetStringArg("inline_assembly_argument_parameter_name")
			inputExecuteAssemblyCommand, _ := taskData.Args.GetStringArg("execute_assembly_command")
			inputExecuteAssemblyFileParameterName, _ := taskData.Args.GetStringArg("execute_assembly_file_parameter_name")
			inputExecuteAssemblyArgumentParameterName, _ := taskData.Args.GetStringArg("execute_assembly_argument_parameter_name")
			remove, _ := taskData.Args.GetBooleanArg("remove_support")
			supportedAgentsFile, err := getOrCreateFile(PayloadTypeSupportFilename)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			newDefinition := agentDefinition{
				Agent:                                inputAgent,
				BofCommand:                           inputBofCommand,
				BofFileParameterName:                 inputBofFileParameterName,
				BofArgumentArrayParameterName:        inputBofArgumentArrayParameterName,
				BofEntryPointParameterName:           inputBofEntrypointParameterName,
				InlineAssemblyCommand:                inputInlineAssemblyCommand,
				InlineAssemblyFileParameterName:      inputInlineAssemblyFileParameterName,
				InlineAssemblyArgumentParameterName:  inputInlineAssemblyArgumentParameterName,
				ExecuteAssemblyCommand:               inputExecuteAssemblyCommand,
				ExecuteAssemblyFileParameterName:     inputExecuteAssemblyFileParameterName,
				ExecuteAssemblyArgumentParameterName: inputExecuteAssemblyArgumentParameterName,
			}
			supportedAgents := []agentDefinition{}
			err = json.Unmarshal(supportedAgentsFile, &supportedAgents)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			found := false
			for i, agent := range supportedAgents {
				if agent.Agent == newDefinition.Agent {
					supportedAgents[i] = newDefinition
					found = true
					if remove {
						supportedAgents = append(supportedAgents[:i], supportedAgents[i+1:]...)
					}
					break
				}
			}
			if !found {
				supportedAgents = append(supportedAgents, newDefinition)
			}
			supportedAgentBytes, err := json.MarshalIndent(supportedAgents, "", "\t")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			err = os.WriteFile(PayloadTypeSupportFilename, supportedAgentBytes, os.ModePerm)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			Initialize()
			rabbitmq.SyncPayloadData(&payloadDefinition.Name, false)
			if remove {
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("Successfully removed support for %s", inputAgent)),
				})
			} else {
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("Successfully added support for %s", inputAgent)),
				})
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
