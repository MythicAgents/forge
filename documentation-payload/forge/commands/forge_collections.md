+++
title = "forge_collections"
chapter = false
weight = 100
hidden = false
+++

## Summary
List out the available commands for a given collections source. The resulting table in the UI will show if a command is already registered or not and give an option to unregister the command.
 
- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### collectionName

- Description: the name of the collection to query
- Required Value: True  
- Default Value: None  

## Usage

```
forge_collections -collectionName SharpCollection
```

## MITRE ATT&CK Mapping

## Detailed Summary
For a given collection, X, this reads the `X_sources.json` file to populate the data in the UI.


