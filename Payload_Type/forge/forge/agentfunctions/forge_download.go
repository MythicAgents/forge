package agentfunctions

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
)

var assemblyVersions = []string{
	"4.0_Any", "4.0_x64", "4.0_x86",
	"4.5_Any", "4.5_x64", "4.5_x86",
	"4.7_Any", "4.7_x64", "4.7_x86",
}

const rateLimitCount = 5
const rateLimitSleep = 5 * time.Second

func rateLimitLoopFetchURL(req *http.Request) ([]byte, error) {
	client := http.Client{}
	env := os.Environ()
	for _, envVar := range env {
		if strings.HasPrefix(envVar, "GITHUB_TOKEN") {
			envPieces := strings.Split(envVar, "=")
			if len(envPieces) != 2 {
				break
			}
			if len(envPieces[1]) < 10 {
				break
			}
			if req.Header.Get("Authorization") == "" {
				req.Header.Add("Authorization", "Bearer "+strings.Split(envVar, "=")[1])
			}
		}
	}
	for i := 0; i < rateLimitCount; i++ {
		resp, err := client.Do(req)
		if err != nil {
			logging.LogError(err, "failed to make network get request")
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logging.LogError(err, "failed to read response body")
			return nil, err
		}
		if resp.StatusCode == 429 || ((resp.StatusCode == 403 || resp.StatusCode == 401) && strings.Contains(string(body), "rate limit")) {
			logging.LogWarning("Hit rate limit, sleeping then trying again", "url", req.URL.String())
			time.Sleep(rateLimitSleep)
			continue
		}
		if resp.StatusCode == 404 {
			return nil, errors.New("resource not found, 404")
		}
		if resp.StatusCode != 200 {
			logging.LogError(nil, "bad status code", "url", req.URL.String(), "status code", resp.StatusCode, "body", body)
			return nil, fmt.Errorf("failed to download repository at %s due to error code: %d", req.URL.String(), resp.StatusCode)
		}
		if resp.StatusCode == 200 {
			return body, nil
		}
	}
	return nil, errors.New("failed to download due to rate limiting")
}
func ExtractTarGz(gzipStream io.Reader, extractPath string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		logging.LogError(err, "ExtractTarGz: NewReader failed")
		return err
	}
	defer uncompressedStream.Close()
	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logging.LogError(err, "ExtractTarGz: Next() failed")
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			logging.LogInfo("extracting", "header folder", header.Name, "extract path", extractPath)
			newName := strings.Replace(header.Name, "./", extractPath, 1)
			err = os.MkdirAll(newName, os.ModePerm)
			if err != nil {
				logging.LogError(err, "ExtractTarGz: Mkdir() failed")
				return err
			}
		case tar.TypeReg:
			logging.LogInfo("extracting", "header files", header.Name, "extract path", extractPath)
			newName := strings.Replace(header.Name, "./", extractPath, 1)
			outFile, err := os.Create(newName)
			if err != nil {
				logging.LogError(err, "ExtractTarGz: Create() failed")
				return err
			}
			_, err = io.Copy(outFile, tarReader)
			outFile.Close()
			if err != nil {
				logging.LogError(err, "ExtractTarGz: Copy() failed")
				return err
			}
		default:
			logging.LogError(nil,
				"ExtractTarGz", "unknown type", header.Typeflag,
				"filename", header.Name)
			return errors.New("unknown tar file type")
		}
	}
	return nil
}
func downloadAssemblyFile(commandSource collectionSourceCommandData, assemblyVersion string, collectionSourceData collectionSource, taskData *agentstructs.PTTaskMessageAllData) error {
	url := fmt.Sprintf("%s/raw/refs/heads/master/NetFramework_%s/%s.exe",
		commandSource.RepoURL, assemblyVersion, commandSource.Name)
	if commandSource.CustomDownloadURL != "" {
		url = commandSource.CustomDownloadURL
	}
	downloadPath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, assemblyVersion, commandSource.Name+".exe")
	err := os.MkdirAll(filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, assemblyVersion),
		os.ModePerm)
	if err != nil {
		return err
	}
	downloadFile, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	if commandSource.customAssemblyFileID == "" {
		if !strings.HasPrefix(url, "http") {
			os.Remove(downloadPath)
			logging.LogError(nil, "no valid http scheme for downloading the file", "url", url)
			return errors.New("no remote url address specified for this command and file missing from disk")
		}
		if taskData != nil {
			mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: []byte(fmt.Sprintf("[*] Downloading %s - v%s...\n", commandSource.Name+".exe", assemblyVersion)),
			})
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logging.LogError(err, "failed to make get request for bof")
			return err
		}

		if taskData != nil {
			if _, ok := taskData.Secrets["GITHUB_TOKEN"]; ok {
				req.Header.Add("Authorization", "Bearer "+taskData.Secrets["GITHUB_TOKEN"].(string))
			}
		}
		body, err := rateLimitLoopFetchURL(req)
		if err != nil {
			if taskData != nil {
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("[!] Failed to download file %s - v%s\n", commandSource.Name+".exe", assemblyVersion)),
				})
			}
			downloadFile.Close()
			os.Remove(downloadPath)
			return err
		}
		_, err = downloadFile.Write(body)
		if err != nil {
			downloadFile.Close()
			os.Remove(downloadPath)
			return err
		}
		if taskData != nil {
			mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: []byte(fmt.Sprintf("[+] Finished Downloading %s - v%s\n", commandSource.Name+".exe", assemblyVersion)),
			})
		}

	} else {
		if taskData != nil {
			mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: []byte(fmt.Sprintf("[*] Fetching %s - v%s from Mythic...\n", commandSource.Name+".exe", assemblyVersion)),
			})
		}

		fileContentsResp, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
			AgentFileID: commandSource.customAssemblyFileID,
		})
		if err != nil {
			logging.LogError(err, "failed to send request to Mythic for contents of file")
			return err
		}
		if !fileContentsResp.Success {
			logging.LogError(errors.New(fileContentsResp.Error), "failed to get file from mythic")
			return errors.New(fileContentsResp.Error)
		}
		_, err = downloadFile.Write(fileContentsResp.Content)
		if err != nil {
			logging.LogError(err, "failed to write contents to disk")
			return err
		}
		downloadFile.Close()
		if taskData != nil {
			mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: []byte(fmt.Sprintf("[*] Saved %s - v%s from Mythic to disk...\n", commandSource.Name+".exe", assemblyVersion)),
			})
		}
	}
	return nil
}
func createAssemblyCommand(commandSource collectionSourceCommandData, collectionSourceData collectionSource, addCommandToFile bool) agentstructs.Command {
	originatingSource := commandSource.RepoURL
	if commandSource.CustomDownloadURL != "" {
		originatingSource = commandSource.CustomDownloadURL
	}
	defaultVersion := "4.7_Any"
	if commandSource.CustomVersion != "" {
		defaultVersion = commandSource.CustomVersion
	}
	defaultChoices := assemblyVersions
	if commandSource.CustomVersion != "" {
		defaultChoices = []string{commandSource.CustomVersion}
	}
	newCommand := agentstructs.Command{
		Name:                fmt.Sprintf("%s%s", AssemblyPrefix, commandSource.CommandName),
		Description:         fmt.Sprintf("%s\nFrom: %s", commandSource.Description, originatingSource),
		HelpString:          fmt.Sprintf("%s%s", AssemblyPrefix, commandSource.CommandName),
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:        []string{agentstructs.SUPPORTED_OS_WINDOWS},
			CommandIsSuggested: true,
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "args",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Arguments to pass to the assembly",
				DefaultValue:     "",
				ModalDisplayName: "Argument String",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     0,
					},
				},
			},
			{
				Name:             "version",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE_CUSTOM,
				Choices:          defaultChoices,
				Description:      "Specify which version of the file to execute",
				DefaultValue:     defaultVersion,
				ModalDisplayName: "Binary Version",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "execution",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:          []string{"inline_assembly", "execute_assembly"},
				Description:      "Specify how the assembly should execute. Execute_assembly is a fork-and-run style architecture, inline_assembly is within the current process.",
				DefaultValue:     "execute_assembly",
				ModalDisplayName: "Execution Options",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			binaryFileID := ""
			arguments, err := taskData.Args.GetStringArg("args")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			assemblyVersion, err := taskData.Args.GetChooseOneArg("version")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			executionMethod, err := taskData.Args.GetChooseOneArg("execution")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			displayParams := fmt.Sprintf("-args \"%s\" -version %s -execution %s", arguments, assemblyVersion, executionMethod)
			response.DisplayParams = &displayParams
			downloadPath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, assemblyVersion, commandSource.Name+".exe")
			downloadFile, err := os.ReadFile(downloadPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// file doesn't exist on disk, try to fetch it first
					fileContents, err := getOrCreateFile(collectionSourceData.SourceFilename)
					if err != nil {
						response.Success = false
						response.Error = err.Error()
						return response
					}
					commandSources := []collectionSourceCommandData{}
					err = json.Unmarshal(fileContents, &commandSources)
					if err != nil {
						response.Success = false
						response.Error = err.Error()
						return response
					}
					foundCommand := false
					for i, _ := range commandSources {
						if commandSources[i].CommandName == commandSource.CommandName {
							foundCommand = true
							updatedStatus := fmt.Sprintf("Downloading assembly...")
							mythicrpc.SendMythicRPCTaskUpdate(mythicrpc.MythicRPCTaskUpdateMessage{
								TaskID:       taskData.Task.ID,
								UpdateStatus: &updatedStatus,
							})
							err = downloadAssemblyFile(commandSources[i], assemblyVersion, collectionSourceData, taskData)
							if err != nil {
								response.Success = false
								response.Error = err.Error()
								return response
							}
						}
					}
					if !foundCommand {
						response.Success = false
						response.Error = fmt.Sprintf("Could not find the command's binary on disk or in the %s file", collectionSourceData.SourceFilename)
						return response
					}
				} else {
					response.Success = false
					response.Error = err.Error()
					return response
				}
			}
			fileSearch, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
				TaskID:     taskData.Task.ID,
				Filename:   fmt.Sprintf("%s.exe", commandSource.Name),
				MaxResults: 1,
				Comment:    fmt.Sprintf("Community Collection's %s.exe version %s", commandSource.Name, assemblyVersion),
			})
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if !fileSearch.Success {
				response.Success = false
				response.Error = fileSearch.Error
				return response
			}
			if len(fileSearch.Files) == 0 {
				// we need to register it first
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("[*] Registering %s.exe with Mythic...\n", commandSource.Name)),
				})
				uploadResponse, err := mythicrpc.SendMythicRPCFileCreate(mythicrpc.MythicRPCFileCreateMessage{
					TaskID:       taskData.Task.ID,
					Filename:     fmt.Sprintf("%s.exe", commandSource.Name),
					Comment:      fmt.Sprintf("Community Collection's %s.exe version %s", commandSource.Name, assemblyVersion),
					FileContents: downloadFile,
				})
				if err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				}
				if !uploadResponse.Success {
					response.Success = false
					response.Error = uploadResponse.Error
					return response
				}
				binaryFileID = uploadResponse.AgentFileId
			} else {
				binaryFileID = fileSearch.Files[0].AgentFileId
			}
			// get the command we're suppose to issue based on this callback's payload type
			registeredAgents := []agentDefinition{}
			agentFileData, err := os.ReadFile(PayloadTypeSupportFilename)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			err = json.Unmarshal(agentFileData, &registeredAgents)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}

			for _, agent := range registeredAgents {
				if agent.Agent == taskData.PayloadType {
					commandName := agent.ExecuteAssemblyCommand
					commandFileArg := agent.ExecuteAssemblyFileParameterName
					commandArgsArg := agent.ExecuteAssemblyArgumentParameterName
					if executionMethod == "inline_assembly" {
						commandName = agent.InlineAssemblyCommand
						commandFileArg = agent.InlineAssemblyFileParameterName
						commandArgsArg = agent.InlineAssemblyArgumentParameterName
					}
					if commandName == "" {
						response.Success = false
						response.Error = "Current payload type doesn't have a supporting execution mechanism for that option"
						return response
					}
					response.CommandName = &commandName
					response.ReprocessAtNewCommandPayloadType = agent.Agent
					taskData.Args.RemoveArg("args")
					taskData.Args.RemoveArg("version")
					taskData.Args.RemoveArg("execution")
					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          commandFileArg,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_FILE,
						DefaultValue:  binaryFileID,
					})
					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          commandArgsArg,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
						DefaultValue:  arguments,
					})

					mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
						TaskID:   taskData.Task.ID,
						Response: []byte(fmt.Sprintf("[*] Passing execution to %s's \"%s\" command for further processing...\n", agent.Agent, commandName)),
					})
					updatedStatus := fmt.Sprintf("%s preparing task...", agent.Agent)
					mythicrpc.SendMythicRPCTaskUpdate(mythicrpc.MythicRPCTaskUpdateMessage{
						TaskID:       taskData.Task.ID,
						UpdateStatus: &updatedStatus,
					})
					return response
				}
			}
			response.Success = false
			response.Error = "Failed to find matching payload type for this callback when looking for supported agents."
			response.Error += fmt.Sprintf("\nModify the %s file to add support for this callback's payload type.", PayloadTypeSupportFilename)
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
	}
	if !addCommandToFile {
		return newCommand
	}
	assemblyCommandsFile, err := getOrCreateFile(collectionSourceData.CommandsFilename)
	if err != nil {
		logging.LogError(err, "Failed to read assembly commands file")
		return newCommand
	}
	assemblyCommands := []assemblyCommand{}
	err = json.Unmarshal(assemblyCommandsFile, &assemblyCommands)
	if err != nil {
		logging.LogError(err, "failed to parse assembly commands into struct")
		return newCommand
	}
	for i, _ := range assemblyCommands {
		if assemblyCommands[i].CommandName == fmt.Sprintf("%s%s", AssemblyPrefix, commandSource.CommandName) {
			// we already have this command Registered, move along
			return newCommand
		}
	}
	assemblyCommands = append(assemblyCommands, assemblyCommand{
		CommandName:           fmt.Sprintf("%s%s", AssemblyPrefix, commandSource.CommandName),
		CollectionType:        collectionSourceData.Name,
		CollectionCommandName: commandSource.Name,
	})
	newAssemblyCommandsBytes, err := json.MarshalIndent(assemblyCommands, "", "\t")
	if err != nil {
		logging.LogError(err, "failed to marshal assembly commands into JSON")
		return newCommand
	}
	err = os.WriteFile(collectionSourceData.CommandsFilename, newAssemblyCommandsBytes, os.ModePerm)
	if err != nil {
		logging.LogError(err, "failed to write out new commands to file")
	}
	return newCommand
}

func downloadBofFile(commandSource collectionSourceCommandData, collectionSourceData collectionSource, taskData *agentstructs.PTTaskMessageAllData) error {
	if len(commandSource.customBofFileIDs) > 0 {
		extractPath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSource.CommandName) + string(os.PathSeparator)
		err := os.MkdirAll(extractPath, os.ModePerm)
		if err != nil {
			return err
		}
		for _, bofFile := range commandSource.customBofFileIDs {
			searchResp, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
				TaskID:      taskData.Task.ID,
				AgentFileID: bofFile,
			})
			if err != nil {
				logging.LogError(err, "failed to send file search request to Mythic")
				return err
			}
			if !searchResp.Success {
				return errors.New(searchResp.Error)
			}
			contentResp, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
				AgentFileID: bofFile,
			})
			if err != nil {
				logging.LogError(err, "failed to send file content request to Mythic")
				return err
			}
			if !contentResp.Success {
				return errors.New(contentResp.Error)
			}
			filePath := filepath.Join(extractPath, searchResp.Files[0].Filename)
			err = os.WriteFile(filePath, contentResp.Content, os.ModePerm)
			if err != nil {
				logging.LogError(err, "failed to write file to disk")
				return err
			}
		}
		contentResp, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
			AgentFileID: commandSource.customBofExtensionFileID,
		})
		if err != nil {
			logging.LogError(err, "failed to send file content request to Mythic")
			return err
		}
		if !contentResp.Success {
			return errors.New(contentResp.Error)
		}
		filePath := filepath.Join(extractPath, "extension.json")
		err = os.WriteFile(filePath, contentResp.Content, os.ModePerm)
		if err != nil {
			logging.LogError(err, "failed to write file to disk")
			return err
		}
		return nil
	}

	downloadPath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSource.CommandName+".tar.gz")
	extractPath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSource.CommandName) + string(os.PathSeparator)

	err := os.MkdirAll(extractPath, os.ModePerm)
	if err != nil {
		return err
	}
	downloadFile, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer downloadFile.Close()
	if taskData != nil {
		mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
			TaskID:   taskData.Task.ID,
			Response: []byte(fmt.Sprintf("[*] Downloading %s...\n", commandSource.Name+".tar.gz")),
		})
	}

	tarGzURL := ""
	if commandSource.CustomDownloadURL != "" {
		logging.LogInfo("Custom download URL was supplied")
		tarGzURL = commandSource.CustomDownloadURL
	} else {
		logging.LogInfo("Using default download procedure, assuming GitHub")
		// calculate GitHub asset download URL
		urlPieces := strings.Split(commandSource.RepoURL, "/")
		repo := strings.Join(urlPieces[3:], "/")
		url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
		if commandSource.CustomVersion != "" {
			url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, commandSource.CustomVersion)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logging.LogError(err, "failed to make get request for bof")
			return err
		}
		// Add required headers
		req.Header.Add("Accept", "application/vnd.github.v3.text-match+json")
		req.Header.Add("Accept", "application/vnd.github.moondragon+json")
		if taskData != nil {
			if _, ok := taskData.Secrets["GITHUB_TOKEN"]; ok {
				req.Header.Add("Authorization", "Bearer "+taskData.Secrets["GITHUB_TOKEN"].(string))
			}
		}
		body, err := rateLimitLoopFetchURL(req)
		if err != nil {
			return err
		}
		result := make(map[string]interface{})
		err = json.Unmarshal(body, &result)
		if err != nil {
			logging.LogError(err, "failed to unmarshal response body")
			return err
		}
		if len(result["assets"].([]interface{})) == 0 {
			return errors.New("no assets found in GitHub release")
		}
		found := false
		for _, asset := range result["assets"].([]interface{}) {
			if asset.(map[string]interface{})["name"].(string) == commandSource.CommandName+".tar.gz" {
				tarGzURL = asset.(map[string]interface{})["url"].(string)
				found = true
			}
		}
		if !found {
			return errors.New("unable to find command name in assets")
		}
	}

	if tarGzURL == "" {
		return errors.New("no download URL present")
	}

	downloadReq, err := http.NewRequest("GET", tarGzURL, nil)
	if err != nil {
		logging.LogError(err, "failed to make new request for bof in released assets")
		return err
	}
	downloadReq.Header.Add("Accept", "application/octet-stream")
	if taskData != nil {
		if _, ok := taskData.Secrets["GITHUB_TOKEN"]; ok {
			downloadReq.Header.Add("Authorization", "Bearer "+taskData.Secrets["GITHUB_TOKEN"].(string))
		}
	}
	downloadFileBody, err := rateLimitLoopFetchURL(downloadReq)
	if err != nil {
		os.Remove(downloadPath)
		return err
	}
	_, err = downloadFile.Write(downloadFileBody)
	if err != nil {
		os.Remove(downloadPath)
		return err
	}
	if taskData != nil {
		mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
			TaskID:   taskData.Task.ID,
			Response: []byte(fmt.Sprintf("[+] Finished Downloading %s\n", commandSource.Name+".tar.gz")),
		})
	}
	downloadFile.Seek(0, 0)
	err = ExtractTarGz(downloadFile, extractPath)
	if err != nil {
		return err
	}
	return nil
}

type bofCommandDefinitionFiles struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	Path string `json:"path"`
}
type bofCommandDefinitionArguments struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	//Type valid values: b -> file, i -> int or integer, s -> short, z -> string, Z -> wstring
	Type     string      `json:"type"`
	Optional bool        `json:"optional"`
	Default  interface{} `json:"default"`
}
type bofCommandDefinition struct {
	Name            string                          `json:"name"`
	Version         string                          `json:"version"`
	CommandName     string                          `json:"command_name"`
	ExtensionAuthor string                          `json:"extension_author"`
	OriginalAuthor  string                          `json:"original_author"`
	RepoURL         string                          `json:"repo_url"`
	Help            string                          `json:"help"`
	LongHelp        string                          `json:"long_help"`
	DependsOn       string                          `json:"depends_on"`
	Entrypoint      string                          `json:"entrypoint"`
	Files           []bofCommandDefinitionFiles     `json:"files"`
	Arguments       []bofCommandDefinitionArguments `json:"arguments"`
}

func createBofCommand(commandSource collectionSourceCommandData, collectionSourceData collectionSource, addCommandToFile bool) (agentstructs.Command, error) {

	bofCommandFolder := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSource.CommandName)
	bofCommandExtensionFilePath := filepath.Join(bofCommandFolder, "extension.json")
	bofCommandExtensionFile, err := os.ReadFile(bofCommandExtensionFilePath)
	if err != nil {
		logging.LogError(err, "failed to find extension.json file")
		return agentstructs.Command{}, err
	}
	bofCommandExtension := bofCommandDefinition{}
	err = json.Unmarshal(bofCommandExtensionFile, &bofCommandExtension)
	if err != nil {
		logging.LogError(err, "failed to unmarshal extension.json file into struct")
		return agentstructs.Command{}, err
	}
	newCommandParameters := []agentstructs.CommandParameter{}
	for i, arg := range bofCommandExtension.Arguments {
		newType := agentstructs.COMMAND_PARAMETER_TYPE_STRING
		switch arg.Type {
		case "file":
			fallthrough
		case "b":
			newType = agentstructs.COMMAND_PARAMETER_TYPE_FILE
		case "int":
			fallthrough
		case "integer":
			fallthrough
		case "i":
			fallthrough
		case "short":
			fallthrough
		case "s":
			newType = agentstructs.COMMAND_PARAMETER_TYPE_NUMBER
		case "string":
			fallthrough
		case "z":
			fallthrough
		case "wstring":
			fallthrough
		case "Z":
			newType = agentstructs.COMMAND_PARAMETER_TYPE_STRING
		}
		newArg := agentstructs.CommandParameter{
			Name:          arg.Name,
			Description:   arg.Description,
			ParameterType: newType,
			DefaultValue:  arg.Default,
			ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
				{
					ParameterIsRequired: !arg.Optional,
					UIModalPosition:     uint32(i),
				},
			},
		}
		newCommandParameters = append(newCommandParameters, newArg)
	}
	newCommand := agentstructs.Command{
		Name: fmt.Sprintf("%s%s", BofPrefix, bofCommandExtension.CommandName),
		Description: fmt.Sprintf("%s\nFrom: %s\nVersion: %s",
			bofCommandExtension.Help, bofCommandExtension.RepoURL, bofCommandExtension.Version),
		HelpString: bofCommandExtension.LongHelp,
		Version:    1,
		Author: fmt.Sprintf("Original: %s, Extension: %s",
			bofCommandExtension.OriginalAuthor, bofCommandExtension.ExtensionAuthor),
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:        []string{agentstructs.SUPPORTED_OS_WINDOWS},
			CommandIsSuggested: true,
		},
		CommandParameters: newCommandParameters,
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			binaryFileID := ""
			typedArgs := make([][]interface{}, len(bofCommandExtension.Arguments))
			displayParams := ""
			for i, arg := range bofCommandExtension.Arguments {
				switch arg.Type {
				case "file":
					fallthrough
				case "b":
					fileId, err := taskData.Args.GetFileArg(arg.Name)
					if err != nil {
						logging.LogError(err, "failed to get file type arg")
						response.Success = false
						response.Error = err.Error()
						return response
					}
					fileSearchResp, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
						AgentFileID: fileId,
					})
					if err != nil {
						logging.LogError(err, "failed to send request to mythic for file")
						response.Success = false
						response.Error = err.Error()
						return response
					}
					if !fileSearchResp.Success {
						logging.LogError(nil, "failed to get file contents from mythic", "err", fileSearchResp.Error)
						response.Success = false
						response.Error = fileSearchResp.Error
						return response
					}
					typedArgs[i] = []interface{}{"b", base64.StdEncoding.EncodeToString(fileSearchResp.Content)}
					displayParams += fmt.Sprintf("-%s %s ", arg.Name, fileId)
				case "int":
					fallthrough
				case "integer":
					fallthrough
				case "i":
					numArg, err := taskData.Args.GetNumberArg(arg.Name)
					if err != nil {
						logging.LogError(err, "failed to get i type arg")
						response.Success = false
						response.Error = err.Error()
						return response
					}
					typedArgs[i] = []interface{}{"i", int(numArg)}
					displayParams += fmt.Sprintf("-%s %d ", arg.Name, int(numArg))
				case "short":
					fallthrough
				case "s":
					numArg, err := taskData.Args.GetNumberArg(arg.Name)
					if err != nil {
						logging.LogError(err, "failed to get s type arg")
						response.Success = false
						response.Error = err.Error()
						return response
					}
					typedArgs[i] = []interface{}{"s", int(numArg)}
					displayParams += fmt.Sprintf("-%s %d ", arg.Name, int(numArg))
				case "string":
					fallthrough
				case "z":
					stringArg, err := taskData.Args.GetStringArg(arg.Name)
					if err != nil {
						logging.LogError(err, "failed to get z type arg")
						response.Success = false
						response.Error = err.Error()
						return response
					}
					typedArgs[i] = []interface{}{"z", stringArg}
					displayParams += fmt.Sprintf("-%s %s ", arg.Name, stringArg)
				case "wstring":
					fallthrough
				case "Z":
					stringArg, err := taskData.Args.GetStringArg(arg.Name)
					if err != nil {
						logging.LogError(err, "failed to get Z type arg")
						response.Success = false
						response.Error = err.Error()
						return response
					}
					typedArgs[i] = []interface{}{"Z", stringArg}
					displayParams += fmt.Sprintf("-%s %s ", arg.Name, stringArg)
				}
				taskData.Args.RemoveArg(arg.Name)
			}
			response.DisplayParams = &displayParams
			targetFilename := ""
			validArchitectures := make([]string, len(bofCommandExtension.Files))
			for i, f := range bofCommandExtension.Files {
				validArchitectures[i] = f.Arch
				if f.OS != "windows" {
					continue
				}
				if f.Arch == "amd64" && slices.Contains([]string{"x64", "amd64", "x86_64"}, strings.ToLower(taskData.Callback.Architecture)) {
					targetFilename = f.Path
				}
				if f.Arch == "386" && slices.Contains([]string{"x86", "i386"}, strings.ToLower(taskData.Callback.Architecture)) {
					targetFilename = f.Path
				}
			}
			if targetFilename == "" {
				response.Success = false
				response.Error = fmt.Sprintf("Callback architecture, %s, doesn't match any bof supported architectures: %s",
					taskData.Callback.Architecture, strings.Join(validArchitectures, ", "))
				return response
			}
			downloadPath := filepath.Join(".", PayloadTypeName, "collections", collectionSourceData.Name, commandSource.CommandName, targetFilename)
			downloadFile, err := os.ReadFile(downloadPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// file doesn't exist on disk, try to fetch it first
					fileContents, err := getOrCreateFile(collectionSourceData.SourceFilename)
					if err != nil {
						response.Success = false
						response.Error = err.Error()
						return response
					}
					commandSources := []collectionSourceCommandData{}
					err = json.Unmarshal(fileContents, &commandSources)
					if err != nil {
						response.Success = false
						response.Error = err.Error()
						return response
					}
					foundCommand := false
					for i, _ := range commandSources {
						if commandSources[i].CommandName == commandSource.CommandName {
							foundCommand = true
							err = downloadBofFile(commandSources[i], collectionSourceData, taskData)
							if err != nil {
								response.Success = false
								response.Error = err.Error()
								return response
							}
						}
					}
					if !foundCommand {
						response.Success = false
						response.Error = fmt.Sprintf("Could not find the command's binary on disk or in the %s file", collectionSourceData.SourceFilename)
						return response
					}
				} else {
					response.Success = false
					response.Error = err.Error()
					return response
				}
			}
			fileSearch, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
				TaskID:     taskData.Task.ID,
				Filename:   targetFilename,
				MaxResults: 1,
				Comment:    fmt.Sprintf("Community Collection's %s version %s", commandSource.CommandName, targetFilename),
			})
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if !fileSearch.Success {
				response.Success = false
				response.Error = fileSearch.Error
				return response
			}
			if len(fileSearch.Files) == 0 {
				// we need to register it first
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(fmt.Sprintf("[*] Registering %s with Mythic...\n", targetFilename)),
				})
				uploadResponse, err := mythicrpc.SendMythicRPCFileCreate(mythicrpc.MythicRPCFileCreateMessage{
					TaskID:       taskData.Task.ID,
					Filename:     targetFilename,
					Comment:      fmt.Sprintf("Community Collection's %s version %s", commandSource.CommandName, targetFilename),
					FileContents: downloadFile,
				})
				if err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				}
				if !uploadResponse.Success {
					response.Success = false
					response.Error = uploadResponse.Error
					return response
				}
				binaryFileID = uploadResponse.AgentFileId
			} else {
				binaryFileID = fileSearch.Files[0].AgentFileId
			}
			// get the command we're suppose to issue based on this callback's payload type
			registeredAgents := []agentDefinition{}
			agentFileData, err := os.ReadFile(PayloadTypeSupportFilename)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			err = json.Unmarshal(agentFileData, &registeredAgents)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}

			for _, agent := range registeredAgents {
				if agent.Agent == taskData.PayloadType {
					commandName := agent.BofCommand
					commandFileArg := agent.BofFileParameterName
					commandArgsArg := agent.BofArgumentArrayParameterName
					response.CommandName = &commandName
					response.ReprocessAtNewCommandPayloadType = agent.Agent

					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          commandFileArg,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_FILE,
						DefaultValue:  binaryFileID,
					})
					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          commandArgsArg,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_TYPED_ARRAY,
						DefaultValue:  typedArgs,
					})
					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          agent.BofEntryPointParameterName,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
						DefaultValue:  bofCommandExtension.Entrypoint,
					})
					newStdout := fmt.Sprintf("%s final args:\nFile: %s\nTyped Args: %v\nEntrypoint: %s\n",
						commandName, binaryFileID, typedArgs, bofCommandExtension.Entrypoint)
					response.Stdout = &newStdout
					mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
						TaskID:   taskData.Task.ID,
						Response: []byte(fmt.Sprintf("[*] Passing execution to %s's \"%s\" command for further processing...\n", agent.Agent, commandName)),
					})
					updatedStatus := fmt.Sprintf("%s preparing task...", agent.Agent)
					mythicrpc.SendMythicRPCTaskUpdate(mythicrpc.MythicRPCTaskUpdateMessage{
						TaskID:       taskData.Task.ID,
						UpdateStatus: &updatedStatus,
					})
					return response
				}
			}
			response.Success = false
			response.Error = "Failed to find matching payload type for this callback when looking for supported agents."
			response.Error += fmt.Sprintf("\nModify the %s file to add support for this callback's payload type.", PayloadTypeSupportFilename)
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
	}
	if !addCommandToFile {
		return newCommand, nil
	}
	assemblyCommandsFile, err := getOrCreateFile(collectionSourceData.CommandsFilename)
	if err != nil {
		logging.LogError(err, "Failed to read assembly commands file")
		return newCommand, err
	}
	assemblyCommands := []bofCommand{}
	err = json.Unmarshal(assemblyCommandsFile, &assemblyCommands)
	if err != nil {
		logging.LogError(err, "failed to parse assembly commands into struct")
		return newCommand, err
	}
	for i, _ := range assemblyCommands {
		if assemblyCommands[i].CommandName == fmt.Sprintf("%s%s", BofPrefix, commandSource.CommandName) {
			// we already have this command Registered, move along
			return newCommand, nil
		}
	}
	assemblyCommands = append(assemblyCommands, bofCommand{
		CommandName:           fmt.Sprintf("%s%s", BofPrefix, commandSource.CommandName),
		CollectionType:        collectionSourceData.Name,
		CollectionCommandName: commandSource.Name,
	})
	newAssemblyCommandsBytes, err := json.MarshalIndent(assemblyCommands, "", "\t")
	if err != nil {
		logging.LogError(err, "failed to marshal assembly commands into JSON")
		return newCommand, err
	}
	err = os.WriteFile(collectionSourceData.CommandsFilename, newAssemblyCommandsBytes, os.ModePerm)
	if err != nil {
		logging.LogError(err, "failed to write out new commands to file")
		return newCommand, err
	}
	return newCommand, nil
}

func init() {
	agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(agentstructs.Command{
		Name:                fmt.Sprintf("%s_download", PayloadTypeName),
		Description:         "Interact with the forge container to download new commands from the web and register them for use.",
		HelpString:          fmt.Sprintf("%s_download -collectionName SharpCollection -commandName Rubeus", PayloadTypeName),
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{fmt.Sprintf("%s:download", PayloadTypeName)},
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
				Description:      "Specify which command to download",
				ModalDisplayName: "Command Name to Download",
				DefaultValue:     "",
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
			displayParams := fmt.Sprintf("-collectionName %s -commandName %s", collection, commandName)
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
			for _, commandSource := range commandSources {
				if commandSource.Name == commandName {
					switch collectionSourceData.Type {
					case "assembly":
						if commandSource.CustomDownloadURL != "" {
							err = downloadAssemblyFile(commandSource, commandSource.CustomVersion, collectionSourceData, taskData)
							if err != nil {
								response.Success = false
								response.Error = err.Error()
								return response
							}
						} else {
							atLeastOneSuccess := false
							for _, assemblyVersion := range assemblyVersions {
								err = downloadAssemblyFile(commandSource, assemblyVersion, collectionSourceData, taskData)
								if err == nil {
									atLeastOneSuccess = true
								}
							}
							if !atLeastOneSuccess {
								response.Success = false
								response.Error = "Failed to download any version of the tool"
								return response
							}
						}
						mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
							TaskID:   taskData.Task.ID,
							Response: []byte(fmt.Sprintf("Registering new command %s%s\n", AssemblyPrefix, commandSource.CommandName)),
						})
						newCommand := createAssemblyCommand(commandSource, collectionSourceData, true)
						agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(newCommand)
					case "bof":
						err = downloadBofFile(commandSource, collectionSourceData, taskData)
						if err != nil {
							response.Success = false
							response.Error = err.Error()
							return response
						}
						mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
							TaskID:   taskData.Task.ID,
							Response: []byte(fmt.Sprintf("Registering new command %s%s\n", BofPrefix, commandSource.CommandName)),
						})
						newCommand, err := createBofCommand(commandSource, collectionSourceData, true)
						if err != nil {
							response.Success = false
							response.Error = err.Error()
							return response
						}
						agentstructs.AllPayloadData.Get(PayloadTypeName).AddCommand(newCommand)
					default:
					}

					rabbitmq.SyncPayloadData(&payloadDefinition.Name, false)
					response.Success = true
					mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
						TaskID:   taskData.Task.ID,
						Response: []byte(fmt.Sprintf("Command Registered for use!\n")),
					})
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
