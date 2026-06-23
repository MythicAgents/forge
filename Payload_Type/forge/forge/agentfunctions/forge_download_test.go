package agentfunctions

import (
	"reflect"
	"testing"
)

func TestExpandBofCommandDefinitionsInheritsPackageMetadata(t *testing.T) {
	definitions := expandBofCommandDefinitions(bofCommandDefinition{
		Name:            "Netuse (SA)",
		PackageName:     "sa-netuse",
		Version:         "v0.0.28",
		ExtensionAuthor: "moloch--",
		OriginalAuthor:  "TrustedSec",
		RepoURL:         "https://github.com/sliverarmory/CS-Situational-Awareness-BOF",
		Commands: []*bofCommandDefinition{
			{
				CommandName: "sa-netuse-add",
				Help:        "Connect to a network share",
				Entrypoint:  "go",
				Files: []bofCommandDefinitionFiles{
					{OS: "windows", Arch: "amd64", Path: "netuse.x64.o"},
				},
			},
			{
				CommandName: "sa-netuse-list",
				Help:        "List network share connections",
				Entrypoint:  "go",
				Files: []bofCommandDefinitionFiles{
					{OS: "windows", Arch: "386", Path: "netuse.x86.o"},
				},
			},
		},
	})

	if len(definitions) != 2 {
		t.Fatalf("expected 2 command definitions, got %d", len(definitions))
	}
	if definitions[0].CommandName != "sa-netuse-add" {
		t.Fatalf("expected first command name to be inherited child command, got %q", definitions[0].CommandName)
	}
	if definitions[0].Version != "v0.0.28" {
		t.Fatalf("expected version metadata to be inherited, got %q", definitions[0].Version)
	}
	if definitions[0].RepoURL != "https://github.com/sliverarmory/CS-Situational-Awareness-BOF" {
		t.Fatalf("expected repo metadata to be inherited, got %q", definitions[0].RepoURL)
	}
	if definitions[1].ExtensionAuthor != "moloch--" {
		t.Fatalf("expected extension author metadata to be inherited, got %q", definitions[1].ExtensionAuthor)
	}
}

func TestBofCommandNamesFromDefinitionsUsesExpandedCommandNames(t *testing.T) {
	commandNames := bofCommandNamesFromDefinitions([]bofCommandDefinition{
		{CommandName: "sa-netuse-add"},
		{CommandName: "sa-netuse-list"},
		{CommandName: "sa-netuse-delete"},
	}, "sa-netuse")

	expectedCommandNames := []string{
		"forge_bof_sa-netuse-add",
		"forge_bof_sa-netuse-list",
		"forge_bof_sa-netuse-delete",
	}
	if !reflect.DeepEqual(commandNames, expectedCommandNames) {
		t.Fatalf("expected command names %v, got %v", expectedCommandNames, commandNames)
	}
}

func TestGetBofCommandSourceNameUsesPackageNameForBundledCommands(t *testing.T) {
	sourceName := getBofCommandSourceName(bofCommandDefinition{
		PackageName: "sa-netuse",
		Commands: []*bofCommandDefinition{
			{CommandName: "sa-netuse-add"},
			{CommandName: "sa-netuse-list"},
		},
	})

	if sourceName != "sa-netuse" {
		t.Fatalf("expected package name source key, got %q", sourceName)
	}
}
