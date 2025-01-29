+++
title = "forge_support"
chapter = false
weight = 101
hidden = false
+++

## Summary
Add, remove, or update support for a payload type for use with forge.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### payload_type

- Description: Choose which payload type to modify support for  
- Required Value: True  
- Default Value: None

#### bof_command

- Description: Name of the bof command for this agent
- Required Value: True
- Default Value: None

#### bof_file_parameter_name

- Description: Name of the parameter that specifies the bof file UUID
- Required Value: True
- Default Value: None

#### bof_argument_array_parameter_name

- Description: Name of the parameter that specifies array of BOF arguments
- Required Value: True
- Default Value: None

#### bof_entrypoint_parameter_name

- Description: Name of the parameter that specifies the bof entrypoint function name
- Required Value: True
- Default Value: None

#### inline_assembly_command

- Description: Name of the command that runs assemblies inline
- Required Value: True
- Default Value: None

#### inline_assembly_file_parameter_name

- Description: Name of the parameter that specifies the assembly file UUID
- Required Value: True
- Default Value: None

#### inline_assembly_argument_parameter_name

- Description: Name of the parameter that specifies the assembly argument string
- Required Value: True
- Default Value: None

#### execute_assembly_command

- Description: Name of the command that runs assemblies in fork-and-run
- Required Value: True
- Default Value: None

#### execute_assembly_file_parameter_name

- Description: Name of the parameter that specifies the assembly file UUID
- Required Value: True
- Default Value: None

#### execute_assembly_argument_parameter_name

- Description: Name of the parameter that specifies the assembly argument string
- Required Value: True
- Default Value: None

#### remove_support

- Description: Remove this agent from the supported list
- Required Value: True
- Default Value: False

## Usage

```
forge_support
```

## MITRE ATT&CK Mapping

## Detailed Summary

