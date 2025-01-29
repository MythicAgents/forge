+++
title = "forge_register"
chapter = false
weight = 101
hidden = false
+++

## Summary
Register a command from a collection with the assumption that the backing .o and exe files already exist in the container.
If `remove` is specified, then remove that command from this callback and all callbacks.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### collectionName

- Description: The name of the collection that contains this command  
- Required Value: True  
- Default Value: None  

#### commandName

- Description: The name of the specific command to register
- Required Value: True
- Default Value: None

#### remove

- Description: If the command is already registered, remove it from this callback and all callbacks
- Required Value: False
- Default Value: False

## Usage

```
forge_register -collectionName SharpCollection -commandName Rubeus
forge_register -collectionName SharpCollection -commandName Rubeus -remove
```

## MITRE ATT&CK Mapping

## Detailed Summary

