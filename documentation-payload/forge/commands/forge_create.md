+++
title = "forge_create"
chapter = false
weight = 103
hidden = false
+++

## Summary
Create an entirely new command by uploading your own BOFs, extension.json, or .NET files. This can be as part of a new "collection" or an existing one.
If there's something in a collection's source of available commands already, you can simply register or download it for use within your callbacks.
This is specifically for uploading your own local data.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### collectionName

- Description: The name of the collection that this command belongs to
- Required Value: True  
- Default Value: 

#### commandName

- Description: The name of the command to register
- Required Value: True
- Default Value:

#### description

- Description: The description of the command
- Required Value: False
- Default Value:

#### commandFileAssembly

- Description: If creating an assembly command, this is the actual exe to run
- Required Value: True
- Default Value:

#### commandVersion

- Description: If creating an assembly command, this is the .net version of the exe
- Required Value: True
- Default Value:

#### commandFilesBof

- Description: If creating a bof command, these are all the .o files to use
- Required Value: True
- Default Value:

#### extensionFile

- Description: If creating a bof command, this is the extension.json file that describes the bof arguments
- Required Value: True
- Default Value:

## Usage

```
forge_create
```


## Detailed Summary
