+++
title = "forge"
chapter = false
weight = 5
+++
![logo](/agents/forge/forge.svg?width=200px)
## Summary

Forge is a "Command Augmentation" payload type that provides a few functions for downloading/creating BOF/.NET commands as new, tab-completable commands within a variety of other payload type's callbacks.
Forge itself can't be "built"; instead, it offers Mythic-side commands that can then be passed down to various callbacks for execution.

These forge commands are automatically injected into all Windows callbacks based on payload types listed in the [payload_type_support.json](#payload_type_supportjson) file (more on that further down).

Forge offers three kinds of execution for supported payload types:
* BOF
* Execute Assembly
* Inline Assembly

The forge container comes with @Flangvik's [SharpCollection](https://github.com/Flangvik/SharpCollection) and Sliver's [Armory](https://github.com/sliverarmory/armory/blob/master/armory.json) installed.


## Supporting Files

### payload_type_support.json

This file identifies which payload types are supported and how to pass execution from forge to them.
This takes the following format:
```json
[
  {
    "agent": "apollo",
    
    "bof_command": "execute_coff",
    "bof_file_parameter_name": "bof_file",
    "bof_argument_array_parameter_name": "coff_arguments",
    "bof_entrypoint_parameter_name": "function_name",

    "inline_assembly_command": "inline_assembly",
    "inline_assembly_file_parameter_name": "assembly_file",
    "inline_assembly_argument_parameter_name": "assembly_arguments",

    "execute_assembly_command": "execute_assembly",
    "execute_assembly_file_parameter_name": "assembly_file",
    "execute_assembly_argument_parameter_name": "assembly_arguments"
  }
]
```
This is an array of entries, one for each agent that's supported. If you want to add your agent as a supported agent, you can use the `forge_support` command. Alternatively, you can modify this file to add your own entry to the list. Then run the following commands:
```bash
sudo ./mythic-cli build forge
```

This file allows you to indicate, for each supported payload type, what parameter names things should get passed down as.

#### bof

BOF commands created as part of Forge are first-order commands within supported callbacks. For example, if the bof command is "sa-netgroup", then the corresponding command that will be registered is `forge_bof_sa-netgroup`.
This allows you to easily group `forge*` commands together and identify if a command is a BOF or .NET executable. If this BOF takes two arguments, `group` and `server`, then they'll be exposed to the operator like 
`forge_bof_sa-netgroup -server 127.0.0.1 -group Administrators`. You don't need to worry about the order or types of values, that'll be handled for you based on the backing bof's `extension.json` file (the same as the SliverArmory format).

This command then needs to be passed down to your callback for your payload type to actually execute the BOF. There are four fields that help identify how this works in your payload_type_support.json:
* "bof_command": "execute_coff"
  * which command in your agent should we pass control to. Control goes right to that command's `create_go_tasking` function and continues from there like normal. 
* "bof_file_parameter_name": "bof_file"
  * every bof is backed by a file, this is identifying the name of your BOF command's parameter that expects the file. The value passed to this parameter is the file's UUID (just like if it was type File)
* "bof_argument_array_parameter_name": "coff_arguments"
  * in Mythic, the ideal way to handle a BOF's arguments is with a parameter type of TypedArray. This allows us to specify the type of data (int, short, widestring, string, etc) in addition to the data itself in a standard format. Data will be passed along in this stnadard format with standard BOF notations.
    * For example: `[ ["i", 5], ["Z", "Administrator"] ]`. Your command's typed array parser will get called with this data.
* "bof_entrypoint_parameter_name": "function_name"
  * BOFs identy the entrypoint for the program. The vast majority of the time this is `go`, but doesn't technically have to be. This is passed into your payload type's command and fetched from this BOF's backing `extension.json` file.

#### net

.NET commands created as part of Forge are first-order commands within supported callbacks. For example, if the .NET is "Rubeus", then the corresponding command that will be registered is `forge_net_Rubeus`.
This allows you to easily group `forge*` commands together and identify if a command is a BOF or .NET executable. Due to the nature of how assemblies are largely used within the offensive community,
these new commands take in a few specific parameters:

* args 
  * this is the entire args string that's passed to the assembly. We can't easily break this up into named parameters because most assemblies just expect a single string as if they were run on the command line.
* version
  * this is the version of the assembly you want to execute. This defaults to `4.7_Any`, but you can set it to any of the versions associated with @Flangvik's SharpCollection repository.
* execution
  * This identifies the execution method you want to use with the assembly - execute_assembly (fork-and-run) or inline_assembly (inside your process)

### collection_sources.json

This file identifies all the collections that are available along with what kind of commands they are. The initial file looks like this:

```json
[
	{
		"name": "SharpCollection",
		"type": "assembly"
	},
	{
		"name": "SliverArmory",
		"type": "bof"
	}
]
```

This tells forge that there's two collections; `SharpCollection` which has `assembly` commands and `SliverArmory` which has `bof` commands.
Forge processes this file and then looks for their associated sources files, `SharpCollection_sources.json` and `SliverArmory_sources.json` files respectively.

### *_sources.json

This file outlines the original sources of all the commands that are available under a specific collection. This is an array of entries, where each one has the following fields:
```json
  {
    "name": "Nano Dump",
    "command_name": "nanodump",
    "description": "",
    "repo_url": "https://github.com/sliverarmory/nanodump",
    "custom_download_url": "", 
    "custom_version": ""
  }
```
* "name":
  * This is the name of the command overall (see this with `forge_community` command)
* "command_name":
  * This is the name of the command that gets added to `forge_bof_` or `forge_net_`
* "description":
  * This is the description of the command and is displayed in the modal UI and help
* "repo_url":
  * assemblies
    * This points to a SharpCollection-like repository for assemblies
  * bof
    * repo that has a tagged release with a command.tar.gz file that contains the extension.json and .o files
* "custom_download_url":
  * This can be the url of a specific file to download instead of using the SharpCollection or SliverArmory release formats
  * assemblies
    * This points to the specific download url of the .exe file
  * bof
    * This points to a specific download url of the command.tar.gz file with the extension.json and .o files
* "custom_version":
  * This can be used to specify a custom version to associate with a .NET execution instead of using one of the versions associated with SharpCollection's formats

If you add your own command sources for an internal repository or download link, you can set a user secret on your account for `GITHUB_TOKEN` with a GitHub pat or any value that you want to use as part of an Authorization header for access. The code to download from the remote repository does the following:
```go
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
```
## Authors
- @its_a_feature_
