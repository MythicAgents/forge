# Implementation Documentation: AdaptixC2 Extension-Kit in Forge (Mythic)

This document provides a comprehensive breakdown of the architecture, implementation steps, file structure, command catalog, and usage instructions for integrating the **full AdaptixC2 Extension-Kit** Beacon Object Files (BOFs) into the **Forge** command augment payload type within **Mythic C2**.

---

## 1. Overview & Architecture

### 1.1 What is Forge in Mythic?
**Forge** is a `Command Augment` payload type in Mythic C2 (`agentstructs.AgentTypeCommandAugment`). It does not build standalone executables to be deployed on targets. Instead, it provides dynamically registered alias commands (prefixed with `forge_bof_` and `forge_net_`) that wrap and pass execution down to supporting agent payload types (such as **Apollo**, **Athena**, **Xenon**, **Poopsie**, **Starburst**).

### 1.2 Full AdaptixC2 Extension-Kit Suite
The Extension-Kit contains **176 BOF modules** spanning 10 primary categories and 5 Active Directory sub-frameworks:
- **Base Modules**: SAL-BOF, SAR-BOF, Elevation-BOF, Creds-BOF, Execution-BOF, Injection-BOF, LateralMovement-BOF, Postex-BOF, Process-BOF, AD-BOF Core.
- **Active Directory Submodules**:
  - **Kerbeus-BOF**: Kerberos authentication, ticket operations, and delegation abuse.
  - **ADCS-BOF**: Active Directory Certificate Services enumeration, request, authentication (PKINIT), and Shadow Credentials.
  - **RelayInformer-BOF**: Security enforcement checks (SMB, MSSQL, HTTP, LDAP signing and channel binding).
  - **MSSQL-BOF**: Complete Microsoft SQL server interaction, query, and privilege escalation suite.
  - **LDAP-BOF**: Full Active Directory LDAP querying, object creation, modification, ACLs, and delegation management.

---

## 2. Integrated Command Catalog (176 Total Commands)

### 2.1 Base Modules (72 Commands)
- **SAL-BOF (25 commands)**: `arp`, `cacls`, `ek-dir`, `ek-env`, `ek-ipconfig`, `listdns`, `ek-netstat`, `ek-nslookup`, `routeprint`, `ek-uptime`, `useridletime`, `ek-whoami`, `alwayselevated`, `hijackablepath`, `tokenpriv`, `unattendfiles`, `unquotedsvc`, `vulndrivers`, `ek-autologon`, `ek-credmanager`, `modautorun`, `modsvc`, `pshistory`, `uacstatus`, `privcheck_all`.
- **SAR-BOF (4 commands)**: `smartscan`, `taskhound`, `quser`, `nbtscan`.
- **Elevation-BOF (5 commands)**: `getsystem_token`, `uac_sspi`, `uac_regshellcmd`, `DCOMPotato`, `printspoofer`.
- **Creds-BOF (9 commands)**: `askcreds`, `cookie-monster`, `get-netntlm`, `ek-hashdump`, `lsadump_secrets`, `lsadump_sam`, `lsadump_cache`, `nanodump`, `underlaycopy`.
- **Execution-BOF (2 commands)**: `ek-execute-assembly`, `noconsolation`.
- **Injection-BOF (4 commands)**: `inject-cfg`, `inject-sec`, `inject-poolparty`, `inject-32to64`.
- **LateralMovement-BOF (7 commands)**: `ek-psexec`, `ek-scshell`, `ek-winrm`, `token_make`, `token_steal`, `runas-user`, `runas-session`.
- **Postex-BOF (3 commands)**: `firewallrule_add`, `screenshot_bof`, `sauroneye`.
- **Process-BOF (5 commands)**: `findmodule`, `findprochandle`, `psc`, `procfreeze_freeze`, `procfreeze_unfreeze`.
- **AD-BOF Core (8 commands)**: `adwssearch`, `badtakeover`, `dcsync-single`, `dcsync-all`, `ek-ldapsearch`, `readlaps`, `webdav_enable`, `webdav_status`.

### 2.2 Active Directory Submodules (104 Commands)
- **Kerbeus-BOF (16 commands)**:
  - `kerbeus_asktgt`, `kerbeus_asktgs`, `kerbeus_asreproasting`, `kerbeus_changepw`, `kerbeus_describe`, `kerbeus_dump`, `kerbeus_hash`, `kerbeus_kerberoasting`, `kerbeus_klist`, `kerbeus_ptt`, `kerbeus_purge`, `kerbeus_renew`, `kerbeus_s4u`, `kerbeus_cross_s4u`, `kerbeus_tgtdeleg`, `kerbeus_triage`.
- **ADCS-BOF (5 commands)**:
  - `certi_auth`, `certi_enum`, `certi_request`, `certi_request_on_behalf`, `certi_shadow`.
- **RelayInformer-BOF (4 commands)**:
  - `relay_http`, `relay_ldap`, `relay_mssql`, `relay_smb`.
- **MSSQL-BOF (28 commands)**:
  - `mssql_1434udp`, `mssql_adsi`, `mssql_agentcmd`, `mssql_agentstatus`, `mssql_checkrpc`, `mssql_clr`, `mssql_columns`, `mssql_databases`, `mssql_disableclr`, `mssql_disableole`, `mssql_disablerpc`, `mssql_disablexp`, `mssql_enableclr`, `mssql_enableole`, `mssql_enablerpc`, `mssql_enablexp`, `mssql_impersonate`, `mssql_info`, `mssql_links`, `mssql_olecmd`, `mssql_query`, `mssql_rows`, `mssql_search`, `mssql_smb`, `mssql_tables`, `mssql_users`, `mssql_whoami`, `mssql_xpcmd`.
- **LDAP-BOF (51 commands)**:
  - `ldap_get-users`, `ldap_get-computers`, `ldap_get-groups`, `ldap_get-usergroups`, `ldap_get-groupmembers`, `ldap_get-object`, `ldap_get-domaininfo`, `ldap_get-maq`, `ldap_get-writable`, `ldap_get-delegation`, `ldap_get-uac`, `ldap_get-attribute`, `ldap_get-spn`, `ldap_get-acl`, `ldap_get-rbcd`, `ldap_add-user`, `ldap_add-computer`, `ldap_add-group`, `ldap_add-groupmember`, `ldap_add-ou`, `ldap_add-sidhistory`, `ldap_add-spn`, `ldap_add-attribute`, `ldap_add-uac`, `ldap_add-delegation`, `ldap_add-rbcd`, `ldap_add-ace`, `ldap_set-password`, `ldap_set-spn`, `ldap_set-delegation`, `ldap_set-attribute`, `ldap_set-uac`, `ldap_set-owner`, `ldap_move-object`, `ldap_remove-groupmember`, `ldap_remove-object`, `ldap_remove-delegation`, `ldap_remove-spn`, `ldap_remove-attribute`, `ldap_remove-rbcd`, `ldap_remove-ace`, `ldap_remove-uac`, and more.

---

## 3. Verification & Usage

### 3.1 Verify Unit Tests
```bash
cd /workspace/forge/Payload_Type/forge
GOROOT=/root/.asdf/installs/golang/1.24.4/go /root/.asdf/installs/golang/1.24.4/go/bin/go test -v ./...
```

### 3.2 Mythic Commands
- Query: `forge_collections -collectionName ExtensionKit`
- Register: `forge_register -collectionName ExtensionKit -commandName kerbeus_asktgt`
- Execute: `forge_bof_kerbeus_asktgt -params "/user:Admin /password:Password123"`
