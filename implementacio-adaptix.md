# Implementation Documentation: AdaptixC2 Extension-Kit in Forge (Mythic)

This document provides a comprehensive breakdown of the architecture, implementation steps, file structure, command catalog, and usage instructions for integrating the **AdaptixC2 Extension-Kit** Beacon Object Files (BOFs) into the **Forge** command augment payload type within **Mythic C2**.

---

## 1. Overview & Architecture

### 1.1 What is Forge in Mythic?
**Forge** is a `Command Augment` payload type in Mythic C2 (`agentstructs.AgentTypeCommandAugment`). It does not build standalone executables to be deployed on targets. Instead, it provides dynamically registered alias commands (prefixed with `forge_bof_` and `forge_net_`) that wrap and pass execution down to supporting agent payload types (such as **Apollo**, **Athena**, **Xenon**, **Poopsie**, **Starburst**).

When an operator issues a `forge_bof_<command>` task in Mythic:
1. Forge parses the user arguments according to the command's `extension.json` descriptor.
2. Forge packages the parameters into a typed argument array.
3. Forge selects the binary matching the target callback architecture (`amd64` / `.x64.o` or `386` / `.x86.o`).
4. Forge registers the binary with Mythic file storage (if not already cached) and delegates the task to the backing agent's native BOF execution command (e.g., `execute_coff` for Apollo or `coff` for Athena).

### 1.2 What is the AdaptixC2 Extension-Kit?
The **Extension-Kit** is an open-source repository containing over 70 BOF modules originally built for AdaptixC2. In AdaptixC2, each module is accompanied by an **AxScript** (`.axs`) file that defines:
- The command name and description via `ax.create_command`.
- Command arguments (strings, integers, wide strings, files) via `cmd.addArg*`.
- A pre-execution hook (`setPreHook`) that packages parameters into a binary format with `ax.bof_pack("types", [args])` and executes the target `.o` file.

### 1.3 How the Integration Works
To bridge AdaptixC2 BOFs into Mythic Forge:
1. **Compilation**: All C/C++ source code in the Extension-Kit subdirectories is compiled using `x86_64-w64-mingw32-gcc`, producing `.x64.o` and `.x86.o` COFF binaries.
2. **Schema Translation (AxScript $\rightarrow$ `extension.json`)**: Each AxScript command and its `ax.bof_pack` type signature are translated into Forge's expected `extension.json` schema.
3. **Type Code Mappings**:
   - `cstr` / ANSI string $\rightarrow$ `z` (`COMMAND_PARAMETER_TYPE_STRING`)
   - `wstr` / UTF-16 wide string $\rightarrow$ `Z` (`COMMAND_PARAMETER_TYPE_STRING`)
   - `int` / `integer` $\rightarrow$ `i` (`COMMAND_PARAMETER_TYPE_NUMBER`)
   - `short` $\rightarrow$ `s` (`COMMAND_PARAMETER_TYPE_NUMBER`)
   - `bytes` / binary file $\rightarrow$ `b` (`COMMAND_PARAMETER_TYPE_FILE`)
4. **Collection Registration**: `ExtensionKit` is registered as a `bof` collection in Forge's `collection_sources.json`, and all 72 commands are cataloged in `ExtensionKit_sources.json`.

---

## 2. Modified Files

### `Payload_Type/forge/collection_sources.json`
The `ExtensionKit` collection source was appended with type `bof`:

```json
[
	{
		"name": "SharpCollection",
		"type": "assembly"
	},
	{
		"name": "SliverArmory",
		"type": "bof"
	},
	{
		"name": "ExtensionKit",
		"type": "bof"
	}
]
```

---

## 3. Created Files

### 3.1 Metadata & State Files (in `Payload_Type/forge/`)
1. **`ExtensionKit_sources.json`**:
   - Master list of all 72 Extension-Kit commands available for inspection and dynamic loading.
2. **`ExtensionKit_commands.json`**:
   - Initially empty `[]`. Tracks active registered commands loaded dynamically by operators.
3. **`generate_extensionkit.py`**:
   - Python automation script that builds directory structures, copies compiled `.o` binaries, and creates all `extension.json` files.

### 3.2 Command Collection Directories (in `Payload_Type/forge/forge/collections/ExtensionKit/`)
72 directories were created, each containing `extension.json` alongside compiled `.x64.o` and `.x86.o` binaries:

```
Payload_Type/forge/forge/collections/ExtensionKit/
├── adwssearch/          ├── ek-whoami/         ├── psc/
├── alwayselevated/      ├── firewallrule_add/  ├── pshistory/
├── arp/                 ├── get-netntlm/       ├── quser/
├── askcreds/            ├── getsystem_token/   ├── readlaps/
├── badtakeover/         ├── hijackablepath/    ├── routeprint/
├── cacls/               ├── inject-32to64/     ├── runas-session/
├── cookie-monster/      ├── inject-cfg/        ├── runas-user/
├── DCOMPotato/          ├── inject-poolparty/  ├── sauroneye/
├── dcsync-all/          ├── inject-sec/        ├── screenshot_bof/
├── dcsync-single/       ├── listdns/           ├── smartscan/
├── ek-autologon/        ├── lsadump_cache/     ├── taskhound/
├── ek-credmanager/      ├── lsadump_sam/       ├── token_make/
├── ek-dir/              ├── lsadump_secrets/   ├── token_steal/
├── ek-env/              ├── modautorun/        ├── tokenpriv/
├── ek-execute-assembly/ ├── modsvc/            ├── uac_regshellcmd/
├── ek-hashdump/         ├── nanodump/          ├── uac_sspi/
├── ek-ipconfig/         ├── nbtscan/           ├── uacstatus/
├── ek-ldapsearch/       ├── noconsolation/     ├── unattendfiles/
├── ek-netstat/          ├── printspoofer/      ├── underlaycopy/
├── ek-nslookup/         ├── privcheck_all/     ├── unquotedsvc/
├── ek-psexec/           ├── procfreeze_freeze/ ├── useridletime/
├── ek-scshell/          ├── procfreeze_unfreeze/├── vulndrivers/
├── ek-uptime/           └── ...                └── webdav_enable/
```

---

## 4. Complete Command Catalog

| Category | Forge Command | Description | Arguments |
| :--- | :--- | :--- | :--- |
| **SAL-BOF** | `arp` | List ARP table | None |
| | `cacls` | Check user permissions on file/folder | `path` (Z) |
| | `ek-dir` | List files in directory (wildcards, recursive) | `directory` (Z), `recursive` (i) |
| | `ek-env` | List process environment variables | None |
| | `ek-ipconfig` | Display IPv4, hostname, DNS server | None |
| | `listdns` | Query and display DNS cache entries | None |
| | `ek-netstat` | Display active network connections | None |
| | `ek-nslookup` | Perform DNS queries | `domain` (z), `type` (z), `server` (z) |
| | `routeprint` | Print IPv4 routing table | None |
| | `ek-uptime` | System boot time and uptime | None |
| | `useridletime` | Display user idle duration | None |
| | `ek-whoami` | Run whoami /all inspection | None |
| | `alwayselevated` | Check AlwaysInstallElevated registry settings | None |
| | `hijackablepath` | Find writable PATH environment directories | None |
| | `tokenpriv` | Inspect token privileges for misconfigurations | None |
| | `unattendfiles` | Look for leftover unattend setup XML files | None |
| | `unquotedsvc` | Enumerate unquoted Windows service binary paths | None |
| | `vulndrivers` | Check for known vulnerable drivers (loldrivers) | None |
| | `ek-autologon` | Check for stored Winlogon registry credentials | None |
| | `ek-credmanager`| Inspect Windows Credential Manager entries | None |
| | `modautorun` | Check for modifiable autorun entries in registry | None |
| | `modsvc` | Check for services with weak DACLs | None |
| | `pshistory` | Inspect PSReadLine history file | None |
| | `uacstatus` | Inspect UAC settings and integrity level | None |
| | `privcheck_all` | Run all privilege checks sequentially | None |
| **SAR-BOF** | `smartscan` | Smart multithreaded port scanner | `target` (z), `scan_level` (i), `custom_ports` (z) |
| | `taskhound` | Collect scheduled tasks from remote systems | `target` (z), `username` (z), `password` (z), `save_directory` (z), `flags` (z) |
| | `quser` | Query remote active user sessions | `host` (z) |
| | `nbtscan` | NetBIOS scanner | `target` (z), `verbose` (i), `quiet` (i), `etc_hosts` (i), `lmhosts` (i), `separator` (z), `timeout` (i) |
| **Elevation-BOF** | `getsystem_token`| Elevate token to SYSTEM / TrustedInstaller | None |
| | `uac_sspi` | UAC bypass via SSPI Datagram Contexts | `path` (z) |
| | `uac_regshellcmd`| UAC bypass via ms-settings shell command | `path` (z) |
| | `DCOMPotato` | LPE via SeImpersonate / DCOM | `use_token` (i), `program` (Z) |
| | `printspoofer` | LPE via Print Spooler named pipe impersonation | `use_token` (i), `program` (Z) |
| **Creds-BOF** | `askcreds` | Prompt user for credentials via dialog | `prompt` (Z), `note` (Z), `wait_time` (i) |
| | `cookie-monster`| Extract browser cookies and encryption keys | `browser` (z), `profile` (z), `browser_pid` (i), flags (i) |
| | `get-netntlm` | Retrieve NetNTLM hash via Internal Monologue | `no_ess` (i) |
| | `ek-hashdump` | Dump SAM account hashes | None |
| | `lsadump_secrets`| Dump LSA secrets from SECURITY hive | None |
| | `lsadump_sam` | Dump SAM account hashes via registry | None |
| | `lsadump_cache` | Dump cached domain credentials (DCC2) | None |
| | `nanodump` | Syscall-based LSASS minidump | `dump_path` (z), flags (i), `chunk_size` (i) |
| | `underlaycopy` | Low-level NTFS file copy (MFT / Metadata) | `mode` (z), `source` (z), `destination` (z), `download` (i) |
| **Execution-BOF** | `ek-execute-assembly` | In-process .NET assembly execution | `assembly` (b), `params` (z) |
| | `noconsolation` | In-memory unmanaged PE execution | `payload` (b), `args` (z) |
| **Injection-BOF** | `inject-cfg` | Injection via CFG function pointer hijack | `pid` (i), `shellcode` (b) |
| | `inject-sec` | Injection via section mapping | `pid` (i), `shellcode` (b) |
| | `inject-poolparty`| Injection via Windows Thread Pool techniques | `pid` (i), `shellcode` (b), `technique` (i) |
| | `inject-32to64` | WOW64 32-bit to native 64-bit injection | `pid` (i), `shellcode` (b) |
| **LateralMovement**| `ek-psexec` | Service creation execution on remote host | `target` (z), `binary` (b), `binary_name` (z), `share` (z), `svc_path` (z), `svc_name` (z), `svc_description` (z) |
| | `ek-scshell` | Fileless service path modification execution | `target` (z), `svc_name` (z), `remote_unc_path` (z), `binary` (b) |
| | `ek-winrm` | Remote command execution via WinRM | `target` (Z), `cmd` (Z), `timeout` (i), `background` (i), `username` (Z), `password` (Z) |
| | `token_make` | Create impersonation token from credentials | `username` (Z), `password` (Z), `domain` (Z), `type` (i) |
| | `token_steal` | Steal and impersonate process token | `pid` (i) |
| | `runas-user` | Execute command under alternate credentials | `username` (Z), `password` (Z), `domain` (Z), `command` (Z), `logon_type` (i), `timeout` (i), `no_output` (i), `bypass_uac` (i) |
| | `runas-session` | Cross-session execution via IHxHelpPaneServer | `session_id` (i), `filepath` (Z) |
| **Postex-BOF** | `firewallrule_add`| Add inbound/outbound firewall rule via COM | `direction` (z), `port` (Z), `rulename` (Z), `rulegroup` (Z), `description` (Z) |
| | `screenshot_bof` | Inline memory-based window/screen capture | `note` (z), `pid` (i) |
| | `sauroneye` | Multi-threaded file keyword & regex search | `cmdline` (z), `directories` (z), `filetypes` (z), `keywords` (z), flags (i) |
| **Process-BOF** | `findmodule` | Enumerate processes with specified DLL loaded| `module` (Z) |
| | `findprochandle`| Find processes with open process handle | `proc` (Z) |
| | `psc` | Enumerate processes with active TCP/RDP | None |
| | `procfreeze_freeze`| Freeze process threads via WerFault PPL bypass| `action` (i), `pid` (i) |
| | `procfreeze_unfreeze`| Resume frozen process threads | `action` (i), `pid` (i) |
| **AD-BOF** | `adwssearch` | Active Directory Web Services query | `query` (z), `attributes` (z), `dc` (z), `dn` (z) |
| | `badtakeover` | BadSuccessor dMSA account takeover | `ou` (z), `account` (z), `sid` (z), `dn` (z), `domain` (z) |
| | `dcsync-single` | DCSync password replication for user | `target` (z), `is_dn` (i), `ou_path` (z), `dc_address` (z), `use_ldaps` (i), `only_nt` (i) |
| | `dcsync-all` | DCSync password replication for all accounts | `ou_path` (z), `dc_address` (z), `use_ldaps` (i), `only_nt` (i), `only_users` (i) |
| | `ek-ldapsearch` | LDAP directory search | `query` (Z), `attributes` (z), `count` (i), `scope` (i), `dc` (z), `dn` (z), `ldaps` (i) |
| | `readlaps` | Retrieve LAPS local admin password | `dc` (z), `dn` (z), `searchFilter` (z), `reportTarget` (z) |
| | `webdav_enable` | Enable WebClient service unprivileged | None |
| | `webdav_status` | Check remote WebDAV listener status | `hosts` (z) |

---

## 5. Usage in Mythic C2

### 5.1 Build and Start Forge
```bash
sudo ./mythic-cli build forge
sudo ./mythic-cli start forge
```

### 5.2 List Available Commands in Callback
```text
forge_collections -collectionName ExtensionKit
```

### 5.3 Register a Command
```text
forge_register -collectionName ExtensionKit -commandName arp
```

### 5.4 Execute the Registered Command
```text
forge_bof_arp
```

For commands with parameters:
```text
forge_bof_ek-nslookup -domain corp.local -type A
forge_bof_underlaycopy -mode MFT -source C:\Windows\System32\config\SAM -destination C:\Temp\sam.bak
```
