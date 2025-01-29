package agentfunctions

import (
	"encoding/json"
	"errors"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/utils/sharedStructs"
	"os"
	"path/filepath"
)

const version = "0.0.1"
const CollectionSources = "collection_sources.json"
const PayloadTypeSupportFilename = "payload_type_support.json"
const BofPrefix = "forge_bof_"
const AssemblyPrefix = "forge_net_"
const PayloadTypeName = "forge"

type collectionSource struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	SourceFilename   string `json:"-"`
	CommandsFilename string `json:"-"`
}

var collectionSourceNotFoundError = errors.New("collection source not found")

func getCollectionSourceNameOptions() []string {
	assemblyCommandsFile, err := getOrCreateFile(CollectionSources)
	if err != nil {
		logging.LogError(err, "Failed to read collection sources file")
		return []string{}
	}
	sources := []collectionSource{}
	err = json.Unmarshal(assemblyCommandsFile, &sources)
	if err != nil {
		logging.LogError(err, "failed to parse collection sources")
		return []string{}
	}
	sourceNames := make([]string, len(sources))
	for i, source := range sources {
		sourceNames[i] = source.Name
	}
	return sourceNames
}
func addCollectionSource(source collectionSource) error {
	collectionFile, err := getOrCreateFile(CollectionSources)
	if err != nil {
		logging.LogError(err, "Failed to read collection sources file")
		return err
	}
	sources := []collectionSource{}
	err = json.Unmarshal(collectionFile, &sources)
	if err != nil {
		logging.LogError(err, "failed to parse collection sources")
		return err
	}
	found := false
	for i, _ := range sources {
		if sources[i].Name == source.Name {
			sources[i] = source
			found = true
		}
	}
	if !found {
		sources = append(sources, source)
	}
	collectionFile, err = json.MarshalIndent(sources, "", "\t")
	if err != nil {
		logging.LogError(err, "failed to parse collection sources")
		return err
	}
	err = os.WriteFile(CollectionSources, collectionFile, 0644)
	return err
}
func getCollectionSource(name string) (collectionSource, error) {
	collection := collectionSource{
		Name:             name,
		SourceFilename:   fmt.Sprintf("%s_sources.json", name),
		CommandsFilename: fmt.Sprintf("%s_commands.json", name),
	}
	collectionFile, err := getOrCreateFile(CollectionSources)
	if err != nil {
		logging.LogError(err, "Failed to read collection sources file")
		return collection, err
	}
	sources := []collectionSource{}
	err = json.Unmarshal(collectionFile, &sources)
	if err != nil {
		logging.LogError(err, "failed to parse collection sources")
		return collection, err
	}
	for i, _ := range sources {
		sources[i].SourceFilename = fmt.Sprintf("%s_sources.json", sources[i].Name)
		sources[i].CommandsFilename = fmt.Sprintf("%s_commands.json", sources[i].Name)
	}
	for _, source := range sources {
		if source.Name == name {
			return source, nil
		}
	}
	return collection, collectionSourceNotFoundError
}
func getCollectionSources() []collectionSource {
	sources := []collectionSource{}
	collectionFile, err := getOrCreateFile(CollectionSources)
	if err != nil {
		logging.LogError(err, "Failed to read collection sources file")
		return sources
	}
	err = json.Unmarshal(collectionFile, &sources)
	if err != nil {
		logging.LogError(err, "failed to parse collection sources")
		return sources
	}
	for i, _ := range sources {
		sources[i].SourceFilename = fmt.Sprintf("%s_sources.json", sources[i].Name)
		sources[i].CommandsFilename = fmt.Sprintf("%s_commands.json", sources[i].Name)
	}
	return sources
}
func getOrCreateFile(filename string) ([]byte, error) {
	commandsFileBytes, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		newFile, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		newFile.Write([]byte("[]"))
		newFile.Close()
		return []byte("[]"), nil
	}
	if err != nil {
		return nil, err
	}
	return commandsFileBytes, nil
}

type collectionSourceCommandData struct {
	Name                     string `json:"name"`
	CommandName              string `json:"command_name"`
	Description              string `json:"description"`
	RepoURL                  string `json:"repo_url"`
	CustomDownloadURL        string `json:"custom_download_url"`
	CustomVersion            string `json:"custom_version"`
	customAssemblyFileID     string
	customBofFileIDs         []string
	customBofExtensionFileID string
	Registered               bool   `json:"registered"`
	Downloadable             bool   `json:"downloadable"`
	CollectionName           string `json:"collection_name"`
}
type agentDefinition struct {
	Agent                                string `json:"agent"`
	BofCommand                           string `json:"bof_command"`
	BofFileParameterName                 string `json:"bof_file_parameter_name"`
	BofArgumentArrayParameterName        string `json:"bof_argument_array_parameter_name"`
	BofEntryPointParameterName           string `json:"bof_entrypoint_parameter_name"`
	InlineAssemblyCommand                string `json:"inline_assembly_command"`
	InlineAssemblyFileParameterName      string `json:"inline_assembly_file_parameter_name"`
	InlineAssemblyArgumentParameterName  string `json:"inline_assembly_argument_parameter_name"`
	ExecuteAssemblyCommand               string `json:"execute_assembly_command"`
	ExecuteAssemblyFileParameterName     string `json:"execute_assembly_file_parameter_name"`
	ExecuteAssemblyArgumentParameterName string `json:"execute_assembly_argument_parameter_name"`
}
type bofCommand struct {
	CommandName           string `json:"command_name"`
	CollectionType        string `json:"collection_type"`
	CollectionCommandName string `json:"collection_command_name"`
}
type assemblyCommand struct {
	CommandName           string `json:"command_name"`
	CollectionType        string `json:"collection_type"`
	CollectionCommandName string `json:"collection_command_name"`
}

var payloadDefinition = agentstructs.PayloadType{
	Name:                                   PayloadTypeName,
	FileExtension:                          "bin",
	Author:                                 "@its_a_feature_",
	SupportedOS:                            []string{agentstructs.SUPPORTED_OS_WINDOWS},
	Wrapper:                                false,
	CanBeWrappedByTheFollowingPayloadTypes: []string{},
	SupportsDynamicLoading:                 true,
	Description:                            fmt.Sprintf("A collection of bofs/assemblies and their associated commands to be shared across agents.\nVersion %s\nNeeds Mythic 3.3.0+", version),
	SupportedC2Profiles:                    []string{},
	MythicEncryptsData:                     true,
	AgentType:                              agentstructs.AgentTypeCommandAugment,
	BuildParameters:                        []agentstructs.BuildParameter{},
	BuildSteps:                             []agentstructs.BuildStep{},
	OnContainerStartFunction: func(message sharedStructs.ContainerOnStartMessage) sharedStructs.ContainerOnStartMessageResponse {
		response := sharedStructs.ContainerOnStartMessageResponse{}
		collectionSources := getCollectionSources()
		for _, source := range collectionSources {
			commandsFile, err := getOrCreateFile(source.CommandsFilename)
			if err != nil {
				logging.LogError(err, "Failed to read commands file")
				response.EventLogErrorMessage = "Failed to read commands file"
				return response
			}
			switch source.Type {
			case "assembly":
				registeredCommands := []assemblyCommand{}
				err = json.Unmarshal(commandsFile, &registeredCommands)
				if err != nil {
					logging.LogError(err, "failed to parse assembly commands into struct")
					response.EventLogErrorMessage = "failed to parse assembly commands into struct"
					return response
				}
				sourceCommandFile, err := getOrCreateFile(source.SourceFilename)
				if err != nil {
					logging.LogError(err, "Failed to read commands file")
					response.EventLogErrorMessage = "Failed to read commands file"
					return response
				}
				sourceCommands := []collectionSourceCommandData{}
				err = json.Unmarshal(sourceCommandFile, &sourceCommands)
				if err != nil {
					logging.LogError(err, "failed to parse assembly commands into struct")
					response.EventLogErrorMessage = "failed to parse assembly commands into struct"
					return response
				}
				for _, registeredCommand := range registeredCommands {
					for _, sourceCommand := range sourceCommands {
						if registeredCommand.CollectionCommandName == sourceCommand.Name {
							newCommand := createAssemblyCommand(sourceCommand, source, false)
							agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(newCommand)
						}
					}
				}
			case "bof":
				registeredCommands := []bofCommand{}
				err = json.Unmarshal(commandsFile, &registeredCommands)
				if err != nil {
					logging.LogError(err, "failed to parse assembly commands into struct")
					response.EventLogErrorMessage = "failed to parse assembly commands into struct"
					return response
				}
				sourceCommandFile, err := getOrCreateFile(source.SourceFilename)
				if err != nil {
					logging.LogError(err, "Failed to read commands file")
					response.EventLogErrorMessage = "Failed to read commands file"
					return response
				}
				sourceCommands := []collectionSourceCommandData{}
				err = json.Unmarshal(sourceCommandFile, &sourceCommands)
				if err != nil {
					logging.LogError(err, "failed to parse assembly commands into struct")
					response.EventLogErrorMessage = "failed to parse assembly commands into struct"
					return response
				}
				for _, registeredCommand := range registeredCommands {
					for _, sourceCommand := range sourceCommands {
						if registeredCommand.CollectionCommandName == sourceCommand.Name {
							newCommand, err := createBofCommand(sourceCommand, source, false)
							if err != nil {
								logging.LogError(err, "failed to create bof command")
								continue
							}
							agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(newCommand)
						}
					}
				}
			default:
			}
		}
		return response
	},
	CheckIfCallbacksAliveFunction: func(message agentstructs.PTCheckIfCallbacksAliveMessage) agentstructs.PTCheckIfCallbacksAliveMessageResponse {
		response := agentstructs.PTCheckIfCallbacksAliveMessageResponse{Success: true, Callbacks: make([]agentstructs.PTCallbacksToCheckResponse, 0)}
		return response
	},
}

func Initialize() {
	agentFileData, err := getOrCreateFile(PayloadTypeSupportFilename)
	if err != nil {
		logging.LogError(err, "Failed to read payload file")
	} else {
		supportedAgents := []agentDefinition{}
		err = json.Unmarshal(agentFileData, &supportedAgents)
		if err != nil {
			logging.LogError(err, "failed to parse payload file")
		} else {
			supportedAgentNames := make([]string, len(supportedAgents))
			for i, agent := range supportedAgents {
				supportedAgentNames[i] = agent.Agent
			}
			payloadDefinition.CommandAugmentSupportedAgents = supportedAgentNames
		}
	}
	// do this to pre-load the existing commands before we sync for the first time
	payloadDefinition.OnContainerStartFunction(sharedStructs.ContainerOnStartMessage{})
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddPayloadDefinition(payloadDefinition)
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddIcon(filepath.Join(".", PayloadTypeName, "agentfunctions", PayloadTypeName+".svg"))
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddDarkModeIcon(filepath.Join(".", PayloadTypeName, "agentfunctions", PayloadTypeName+".svg"))
}
