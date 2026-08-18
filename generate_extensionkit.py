#!/usr/bin/env python3
"""
Generate Forge-compatible extension.json files for the full AdaptixC2 Extension-Kit,
including all base modules (SAL, SAR, Elevation, Creds, Execution, Injection,
LateralMovement, Postex, Process, AD) and submodules (Kerbeus, ADCS, RelayInformer,
SQL, LDAP).

Forge argument types (from forge_download.go):
  b/file   -> COMMAND_PARAMETER_TYPE_FILE
  i/int    -> COMMAND_PARAMETER_TYPE_NUMBER
  s/short  -> COMMAND_PARAMETER_TYPE_NUMBER
  z/string -> COMMAND_PARAMETER_TYPE_STRING
  Z/wstring-> COMMAND_PARAMETER_TYPE_STRING
"""

import json
import os
import re
import shutil
from pathlib import Path

# Paths
EXTENSION_KIT_DIR = Path("/workspace/forge/Extension-Kit")
FORGE_COLLECTIONS_DIR = Path("/workspace/forge/Payload_Type/forge/forge/collections/ExtensionKit")
FORGE_BASE_DIR = Path("/workspace/forge/Payload_Type/forge")

BOFPACK_MAP = {
    "bytes": "b",
    "int": "i",
    "short": "s",
    "cstr": "z",
    "wstr": "Z",
}

# ==============================================================================
# Base Commands Registry (Manually Curated)
# ==============================================================================

BASE_COMMANDS = [
    # SAL-BOF
    {"cmd": "arp", "desc": "List ARP table", "bof": "arp", "module": "SAL-BOF", "args": []},
    {"cmd": "cacls", "desc": "List user permissions for file/directory", "bof": "cacls", "module": "SAL-BOF", "args": [("path", "Z", "File or directory path", False, None)]},
    {"cmd": "ek-dir", "desc": "Lists files in directory (wildcards, recursive)", "bof": "dir", "module": "SAL-BOF", "args": [("directory", "Z", "Directory to list", True, ".\\\\"), ("recursive", "i", "Recursive list (1=yes, 0=no)", True, 0)]},
    {"cmd": "ek-env", "desc": "List process environment variables", "bof": "env", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-ipconfig", "desc": "List IPv4 address, hostname, and DNS server", "bof": "ipconfig", "module": "SAL-BOF", "args": []},
    {"cmd": "listdns", "desc": "List DNS cache entries", "bof": "listdns", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-netstat", "desc": "Display network connections", "bof": "netstat", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-nslookup", "desc": "Make a DNS query", "bof": "nslookup", "module": "SAL-BOF", "args": [("domain", "z", "Domain to query", False, None), ("type", "z", "Record type", True, "A"), ("server", "z", "DNS server", True, "")]},
    {"cmd": "routeprint", "desc": "List IPv4 routes", "bof": "routeprint", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-uptime", "desc": "System boot time and uptime", "bof": "uptime", "module": "SAL-BOF", "args": []},
    {"cmd": "useridletime", "desc": "Shows user idle time", "bof": "useridletime", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-whoami", "desc": "List whoami /all", "bof": "whoami", "module": "SAL-BOF", "args": []},
    {"cmd": "alwayselevated", "desc": "Check Always Install Elevated registry setting", "bof": "alwayselevated", "module": "SAL-BOF", "args": []},
    {"cmd": "hijackablepath", "desc": "Check PATH for writable directories", "bof": "hijackablepath", "module": "SAL-BOF", "args": []},
    {"cmd": "tokenpriv", "desc": "List token privileges and vulnerable ones", "bof": "tokenpriv", "module": "SAL-BOF", "args": []},
    {"cmd": "unattendfiles", "desc": "Check for leftover unattend setup files", "bof": "unattendfiles", "module": "SAL-BOF", "args": []},
    {"cmd": "unquotedsvc", "desc": "Check for unquoted service paths", "bof": "unquotedsvc", "module": "SAL-BOF", "args": []},
    {"cmd": "vulndrivers", "desc": "Check for known vulnerable drivers (loldrivers.io)", "bof": "vulndrivers", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-autologon", "desc": "Check for stored Autologon registry credentials", "bof": "autologon", "module": "SAL-BOF", "args": []},
    {"cmd": "ek-credmanager", "desc": "Enumerate Windows Credential Manager entries", "bof": "credmanager", "module": "SAL-BOF", "args": []},
    {"cmd": "modautorun", "desc": "Check for modifiable autorun executables", "bof": "modautorun", "module": "SAL-BOF", "args": []},
    {"cmd": "modsvc", "desc": "Check for services with modifiable permissions (DACL)", "bof": "modsvc", "module": "SAL-BOF", "args": []},
    {"cmd": "pshistory", "desc": "Check for PowerShell PSReadLine history file", "bof": "pshistory", "module": "SAL-BOF", "args": []},
    {"cmd": "uacstatus", "desc": "Check UAC status, integrity level, and local admin group", "bof": "uacstatus", "module": "SAL-BOF", "args": []},
    {"cmd": "privcheck_all", "desc": "Run all privilege escalation checks sequentially", "bof": "privcheck_all", "module": "SAL-BOF", "args": []},

    # SAR-BOF
    {"cmd": "smartscan", "desc": "Smart port scan", "bof": "smartscan", "module": "SAR-BOF", "args": [("target", "z", "IP address, range, or CIDR", False, None), ("scan_level", "i", "Scan level: 1=fast, 2=standard, 3=full, 0=custom", True, 2), ("custom_ports", "z", "Custom ports", True, "")]},
    {"cmd": "taskhound", "desc": "Collect scheduled tasks from remote systems", "bof": "taskhound", "module": "SAR-BOF", "args": [("target", "z", "Remote system IP/hostname", False, None), ("username", "z", "Username", True, ""), ("password", "z", "Password", True, ""), ("save_directory", "z", "Save directory", True, ""), ("flags", "z", "Flags", True, "")]},
    {"cmd": "quser", "desc": "Query user sessions on a remote machine", "bof": "quser", "module": "SAR-BOF", "args": [("host", "z", "Remote host", True, "localhost")]},
    {"cmd": "nbtscan", "desc": "NetBIOS name scanner", "bof": "nbtscan", "module": "SAR-BOF", "args": [("target", "z", "IP address/range/CIDR", False, None), ("verbose", "i", "Verbose (1=yes)", True, 0), ("quiet", "i", "Quiet (1=yes)", True, 0), ("etc_hosts", "i", "etc_hosts (1=yes)", True, 0), ("lmhosts", "i", "lmhosts (1=yes)", True, 0), ("separator", "z", "Separator", True, ""), ("timeout", "i", "Timeout ms", True, 1000)]},

    # Elevation-BOF
    {"cmd": "getsystem_token", "desc": "Elevate to SYSTEM via token impersonation", "bof": "getsystem_token", "module": "Elevation-BOF", "args": []},
    {"cmd": "uac_sspi", "desc": "UAC bypass via SSPI Datagram Contexts", "bof": "uac_sspi", "module": "Elevation-BOF", "args": [("path", "z", "Path to agent binary", False, None)]},
    {"cmd": "uac_regshellcmd", "desc": "UAC bypass via ms-settings registry key", "bof": "uac_regshellcmd", "module": "Elevation-BOF", "args": [("path", "z", "Path to agent binary", False, None)]},
    {"cmd": "DCOMPotato", "desc": "DCOMPotato - get SYSTEM via SeImpersonate privileges", "bof": "DCOMPotato", "module": "Elevation-BOF", "args": [("use_token", "i", "Elevate current agent (1=token, 0=run program)", False, None), ("program", "Z", "Program to run", True, "")]},
    {"cmd": "printspoofer", "desc": "LPE via Print Spooler (Named Pipe Impersonation)", "bof": "printspoofer", "module": "Elevation-BOF", "args": [("use_token", "i", "Elevate current agent (1=token, 0=run program)", False, None), ("program", "Z", "Program to run", True, "")]},

    # Creds-BOF
    {"cmd": "askcreds", "desc": "Prompt for credentials via Windows UI", "bof": "askcreds", "module": "Creds-BOF", "args": [("prompt", "Z", "Dialog title", True, "Restore Network Connection"), ("note", "Z", "Dialog note", True, "Please verify credentials"), ("wait_time", "i", "Wait time (s)", True, 30)]},
    {"cmd": "cookie-monster", "desc": "Locate and copy cookie files for Edge/Chrome/Firefox", "bof": "cookie-monster-bof", "module": "Creds-BOF", "args": [("browser", "z", "Browser name", True, ""), ("profile", "z", "Profile path", True, ""), ("browser_pid", "i", "Browser PID", True, 0), ("dump_cookie", "i", "Dump cookie (1=yes)", True, 0), ("dump_password", "i", "Dump password (1=yes)", True, 0), ("dump_key", "i", "Dump key (1=yes)", True, 0), ("cookie_pid", "i", "Cookie PID", True, 0), ("password_pid", "i", "Password PID", True, 0)]},
    {"cmd": "get-netntlm", "desc": "Retrieve NetNTLM hash for the current user", "bof": "get-netntlm", "module": "Creds-BOF", "args": [("no_ess", "i", "Disable session security for NetNTLMv1 (1=yes)", True, 0)]},
    {"cmd": "ek-hashdump", "desc": "Dump SAM hashes", "bof": "hashdump", "module": "Creds-BOF", "args": []},
    {"cmd": "lsadump_secrets", "desc": "Dump LSA secrets from SECURITY hive (requires SYSTEM)", "bof": "lsadump_secrets", "module": "Creds-BOF", "args": []},
    {"cmd": "lsadump_sam", "desc": "Dump SAM hashes via registry (requires admin)", "bof": "lsadump_sam", "module": "Creds-BOF", "args": []},
    {"cmd": "lsadump_cache", "desc": "Dump cached domain credentials (DCC2, requires SYSTEM)", "bof": "lsadump_cache", "module": "Creds-BOF", "args": []},
    {"cmd": "nanodump", "desc": "Use syscalls to dump LSASS", "bof": "nanodump", "module": "Creds-BOF", "args": [("dump_path", "z", "Dump path", True, ""), ("write_file", "i", "Write to file", True, 0), ("use_valid_sig", "i", "Valid signature", True, 0), ("fork", "i", "Fork process", True, 0), ("snapshot", "i", "Snapshot process", True, 0), ("dup", "i", "Duplicate handle", True, 0), ("elevate_handle", "i", "Elevate handle", True, 0), ("duplicate_elevate", "i", "Duplicate elevate", True, 0), ("use_seclogon_leak_local", "i", "Seclogon leak local", True, 0), ("use_seclogon_leak_remote", "i", "Seclogon leak remote", True, 0), ("seclogon_leak_remote_binary", "z", "Remote binary path", True, ""), ("use_seclogon_duplicate", "i", "Seclogon duplicate", True, 0), ("spoof_callstack", "i", "Spoof callstack", True, 0), ("use_silent_process_exit", "i", "Silent process exit", True, 0), ("silent_process_exit", "z", "Silent exit folder", True, ""), ("use_lsass_shtinkering", "i", "LSASS Shtinkering", True, 0), ("get_pid", "i", "Get PID only", True, 0), ("pid", "i", "Target LSASS PID", True, 0), ("chunk_size", "i", "Chunk size bytes", True, 921600)]},
    {"cmd": "underlaycopy", "desc": "Copy file using low-level NTFS access (MFT/Metadata)", "bof": "underlaycopy", "module": "Creds-BOF", "args": [("mode", "z", "Copy mode: MFT or Metadata", False, None), ("source", "z", "Source file path", False, None), ("destination", "z", "Destination file path", True, ""), ("download", "i", "Download instead (1=yes)", True, 0)]},

    # Execution-BOF
    {"cmd": "ek-execute-assembly", "desc": "Perform in-process .NET assembly execution", "bof": "execute-assembly", "module": "Execution-BOF", "args": [("assembly", "b", ".NET assembly file", False, None), ("params", "z", ".NET parameters", True, "")]},
    {"cmd": "noconsolation", "desc": "Run unmanaged EXE/DLL inside agent memory", "bof": "NoConsolation", "module": "Execution-BOF", "args": [("payload", "b", "EXE/DLL payload file", False, None), ("args", "z", "Arguments string", True, "")]},

    # Injection-BOF
    {"cmd": "inject-cfg", "desc": "Inject shellcode via CFG function pointer overwrite", "bof": "inject_cfg", "module": "Injection-BOF", "args": [("pid", "i", "Target process ID", False, None), ("shellcode", "b", "Shellcode file", False, None)]},
    {"cmd": "inject-sec", "desc": "Inject shellcode via section mapping", "bof": "inject_sec", "module": "Injection-BOF", "args": [("pid", "i", "Target process ID", False, None), ("shellcode", "b", "Shellcode file", False, None)]},
    {"cmd": "inject-poolparty", "desc": "Inject shellcode via Pool Party technique", "bof": "inject_poolparty", "module": "Injection-BOF", "args": [("pid", "i", "Target process ID", False, None), ("shellcode", "b", "Shellcode file", False, None), ("technique", "i", "Technique variant (1-8)", False, None)]},
    {"cmd": "inject-32to64", "desc": "Inject x64 shellcode from WOW64 into native 64-bit process", "bof": "inject_32to64", "module": "Injection-BOF", "args": [("pid", "i", "Target process ID", False, None), ("shellcode", "b", "Shellcode file", False, None)]},

    # LateralMovement-BOF
    {"cmd": "ek-psexec", "desc": "Spawn session on remote target via PsExec", "bof": "psexec", "module": "LateralMovement-BOF", "args": [("target", "z", "Remote target IP/hostname", False, None), ("binary", "b", "Binary file to upload", False, None), ("binary_name", "z", "Remote binary name", True, "random"), ("share", "z", "Share name", True, "ADMIN$"), ("svc_path", "z", "Service path", True, "C:\\Windows"), ("svc_name", "z", "Service name", True, "random"), ("svc_description", "z", "Service description", True, "random")]},
    {"cmd": "ek-scshell", "desc": "Spawn session on remote target via SCShell (fileless)", "bof": "scshell", "module": "LateralMovement-BOF", "args": [("target", "z", "Remote target", False, None), ("svc_name", "z", "Service name", True, "defragsvc"), ("remote_unc_path", "z", "Remote UNC path", True, ""), ("binary", "b", "Binary file", False, None)]},
    {"cmd": "ek-winrm", "desc": "Use WinRM to execute commands on other systems", "bof": "winrm", "module": "LateralMovement-BOF", "args": [("target", "Z", "Remote target", False, None), ("cmd", "Z", "Command line", False, None), ("timeout", "i", "Timeout ms", True, 0), ("background", "i", "Keep shell open", True, 0), ("username", "Z", "Username", True, ""), ("password", "Z", "Password", True, "")]},
    {"cmd": "token_make", "desc": "Create impersonation token from credentials", "bof": "token_make", "module": "LateralMovement-BOF", "args": [("username", "Z", "Username", False, None), ("password", "Z", "Password", False, None), ("domain", "Z", "Domain", False, None), ("type", "i", "Logon type", False, None)]},
    {"cmd": "token_steal", "desc": "Steal access token from a process", "bof": "token_steal", "module": "LateralMovement-BOF", "args": [("pid", "i", "Process ID", False, None)]},
    {"cmd": "runas-user", "desc": "Run command as another user using explicit credentials", "bof": "runas", "module": "LateralMovement-BOF", "args": [("username", "Z", "Username", False, None), ("password", "Z", "Password", False, None), ("domain", "Z", "Domain", False, None), ("command", "Z", "Command line", False, None), ("logon_type", "i", "Logon type", True, 2), ("timeout", "i", "Timeout ms", True, 0), ("no_output", "i", "No output (1=yes)", True, 1), ("bypass_uac", "i", "Bypass UAC", True, 0)]},
    {"cmd": "runas-session", "desc": "Execute binary in another user session via IHxHelpPaneServer", "bof": "runas_sess_ihxexec", "module": "LateralMovement-BOF", "args": [("session_id", "i", "Session ID", False, None), ("filepath", "Z", "File path", False, None)]},

    # Postex-BOF
    {"cmd": "firewallrule_add", "desc": "Add inbound/outbound firewall rule via COM", "bof": "addfirewallrule", "module": "Postex-BOF", "args": [("direction", "z", "Direction: in/out", True, "in"), ("port", "Z", "Port number", False, None), ("rulename", "Z", "Rule name", False, None), ("rulegroup", "Z", "Rule group", True, ""), ("description", "Z", "Rule description", True, "")]},
    {"cmd": "screenshot_bof", "desc": "Inline memory-based screenshot", "bof": "Screenshot", "module": "Postex-BOF", "args": [("note", "z", "Caption", True, "ScreenshotBOF"), ("pid", "i", "Window PID (0=full screen)", True, 0)]},
    {"cmd": "sauroneye", "desc": "Search directories for files with keywords (SauronEye BOF)", "bof": "sauroneye", "module": "Postex-BOF", "args": [("cmdline", "z", "Command line", True, ""), ("directories", "z", "Directories", True, "C:\\"), ("filetypes", "z", "Extensions", True, ".txt,.docx"), ("keywords", "z", "Keywords", True, ""), ("search_contents", "i", "Search contents", True, 0), ("max_filesize", "i", "Max size KB", True, 1024), ("system_dirs", "i", "System dirs", True, 0), ("before_date", "z", "Before date", True, ""), ("after_date", "z", "After date", True, ""), ("check_macro", "i", "Check macros", True, 0), ("show_date", "i", "Show dates", True, 0), ("wildcard_attempts", "i", "Wildcard attempts", True, 1000), ("wildcard_size", "i", "Wildcard size KB", True, 200), ("wildcard_backtrack", "i", "Backtrack limit", True, 1000)]},

    # Process-BOF
    {"cmd": "findmodule", "desc": "Identify processes which have a certain module loaded", "bof": "findmodule", "module": "Process-BOF", "args": [("module", "Z", "Module name", False, None)]},
    {"cmd": "findprochandle", "desc": "Identify processes with a specific process handle", "bof": "findprochandle", "module": "Process-BOF", "args": [("proc", "Z", "Process name", False, None)]},
    {"cmd": "psc", "desc": "Shows processes with established TCP and RDP connections", "bof": "psc", "module": "Process-BOF", "args": []},
    {"cmd": "procfreeze_freeze", "desc": "Freeze target process threads via WerFault PPL bypass", "bof": "procfreeze", "module": "Process-BOF", "args": [("action", "i", "Action (1=freeze)", True, 1), ("pid", "i", "Process ID", False, None)]},
    {"cmd": "procfreeze_unfreeze", "desc": "Unfreeze a previously frozen process", "bof": "procfreeze", "module": "Process-BOF", "args": [("action", "i", "Action (2=unfreeze)", True, 2), ("pid", "i", "Unused", True, 0)]},

    # AD-BOF Core
    {"cmd": "adwssearch", "desc": "Executes ADWS query", "bof": "adws_search", "module": "AD-BOF", "args": [("query", "z", "ADWS filter", False, None), ("attributes", "z", "Attributes", True, ""), ("dc", "z", "Target DC", True, ""), ("dn", "z", "Base DN", True, "")]},
    {"cmd": "badtakeover", "desc": "Account takeover via BadSuccessor technique", "bof": "badtakeover", "module": "AD-BOF", "args": [("ou", "z", "Target OU", False, None), ("account", "z", "New dMSA name", False, None), ("sid", "z", "Current context SID", False, None), ("dn", "z", "Target user DN", False, None), ("domain", "z", "Current domain", False, None)]},
    {"cmd": "dcsync-single", "desc": "Perform DCSync on a single user", "bof": "dcsync-single", "module": "AD-BOF", "args": [("target", "z", "Target user or DN", False, None), ("is_dn", "i", "Is DN (1=yes)", True, 0), ("ou_path", "z", "OU path", True, ""), ("dc_address", "z", "DC address", True, ""), ("use_ldaps", "i", "Use LDAPS", True, 0), ("only_nt", "i", "Only NTLM hashes", True, 0)]},
    {"cmd": "dcsync-all", "desc": "Perform DCSync for all users in the domain", "bof": "dcsync-all", "module": "AD-BOF", "args": [("ou_path", "z", "OU path", True, ""), ("dc_address", "z", "DC address", True, ""), ("use_ldaps", "i", "Use LDAPS", True, 0), ("only_nt", "i", "Only NTLM hashes", True, 0), ("only_users", "i", "Only User/Trust accounts", True, 0)]},
    {"cmd": "ek-ldapsearch", "desc": "Executes LDAP query", "bof": "ldapsearch", "module": "AD-BOF", "args": [("query", "Z", "LDAP filter", False, None), ("attributes", "z", "Attributes", True, "*"), ("count", "i", "Result max size", True, 0), ("scope", "i", "Scope (1-3)", True, 3), ("dc", "z", "DC hostname", True, ""), ("dn", "z", "Base DN", True, ""), ("ldaps", "i", "Use LDAPS", True, 0)]},
    {"cmd": "readlaps", "desc": "Read LAPS password for a computer", "bof": "readlaps", "module": "AD-BOF", "args": [("dc", "z", "Target DC", True, ""), ("dn", "z", "Root DN", True, ""), ("searchFilter", "z", "LDAP search filter", True, ""), ("reportTarget", "z", "Target computer name", True, "")]},
    {"cmd": "webdav_enable", "desc": "Enable WebDAV client service unprivileged", "bof": "webdav_enable", "module": "AD-BOF", "args": []},
    {"cmd": "webdav_status", "desc": "Determine if WebDAV is running on remote system", "bof": "webdav_status", "module": "AD-BOF", "args": [("hosts", "z", "Comma-separated hosts", True, "127.0.0.1")]},
]

# ==============================================================================
# Dynamic Parser for Submodules (Kerbeus, ADCS, RelayInformer, SQL, LDAP)
# ==============================================================================

def parse_submodule_axs(rel_path, prefix, module_name):
    full_path = EXTENSION_KIT_DIR / rel_path
    if not full_path.exists():
        return []

    with open(full_path, "r", encoding="utf-8") as f:
        content = f.read()

    pattern = re.compile(r'(?:var|let)\s+(\w+)\s*=\s*ax\.create_command\(\s*\"([^\"]+)\"\s*,\s*\"([^\"]+)\"')
    matches = list(pattern.finditer(content))

    commands = []
    for i, m in enumerate(matches):
        var_name = m.group(1)
        cmd_name = m.group(2)
        cmd_desc = m.group(3)

        end_pos = matches[i + 1].start() if i + 1 < len(matches) else len(content)
        block = content[m.start():end_pos]

        if "setPreHook" not in block:
            continue

        bof_m = re.search(r'_bin/([^\"]+?)\.', block)
        if not bof_m:
            continue
        bof_rel = bof_m.group(1)
        bof_base = os.path.basename(bof_rel)

        # Parse bof_pack
        pack_m = re.search(r'ax\.bof_pack\(\s*\"([^\"]+)\"', block)
        pack_types = [BOFPACK_MAP.get(t.strip(), "z") for t in pack_m.group(1).split(",")] if pack_m else []

        # Parse arguments
        arg_pats = [
            (re.compile(r'\.addArgString\(\s*\"([^\"]+)\"\s*(?:,\s*(true|false|\"[^\"]*\"))?'), "z"),
            (re.compile(r'\.addArgFlagString\(\s*\"[^\"]+\"\s*,\s*\"([^\"]+)\"\s*(?:,\s*(?:true|false|\"[^\"]*\"))?'), "z"),
            (re.compile(r'\.addArgInt\(\s*\"([^\"]+)\"\s*(?:,\s*(true|false|\d+))?'), "i"),
            (re.compile(r'\.addArgFlagInt\(\s*\"[^\"]+\"\s*,\s*\"([^\"]+)\"\s*(?:,\s*(?:true|false|\d+))?'), "i"),
            (re.compile(r'\.addArgFile\(\s*\"([^\"]+)\"\s*(?:,\s*(true|false))?'), "b"),
            (re.compile(r'\.addArgFlagFile\(\s*\"[^\"]+\"\s*,\s*\"([^\"]+)\"\s*(?:,\s*(true|false))?'), "b"),
        ]

        found_args = []
        for pat, atype in arg_pats:
            for am in pat.finditer(block):
                aname = am.group(1)
                opt = True
                if am.lastindex and am.lastindex >= 2 and am.group(2) == "true":
                    opt = False
                found_args.append((am.start(), aname, atype, opt))

        found_args.sort(key=lambda x: x[0])

        args = []
        for idx, ptype in enumerate(pack_types):
            if idx < len(found_args):
                _, aname, _, opt = found_args[idx]
                args.append((aname, ptype, aname, opt, None))
            else:
                args.append((f"arg{idx+1}", ptype, f"Parameter {idx+1}", True, None))

        final_cmd_name = f"{prefix}_{cmd_name}" if prefix else cmd_name
        commands.append({
            "cmd": final_cmd_name,
            "desc": cmd_desc,
            "bof": bof_base,
            "module": module_name,
            "args": args,
        })

    return commands


def find_bof_binaries(cmd_def):
    module_dir = EXTENSION_KIT_DIR / cmd_def["module"]
    bof_name = cmd_def["bof"]
    found_files = {}

    for root, dirs, files in os.walk(module_dir):
        for f in files:
            if not f.endswith(".o"):
                continue
            path = Path(root) / f
            # Match 64-bit
            if f.lower() == f"{bof_name.lower()}.x64.o" or f.lower() == f"{bof_name.lower()}.x86_64.o":
                found_files["amd64"] = path
            # Match 32-bit
            elif f.lower() == f"{bof_name.lower()}.x32.o" or f.lower() == f"{bof_name.lower()}.x86.o" or f.lower() == f"{bof_name.lower()}.i386.o":
                found_files["386"] = path

    return found_files


def copy_and_generate_files(cmd_def, dest_dir):
    found_binaries = find_bof_binaries(cmd_def)
    bof_name = cmd_def["bof"]
    file_entries = []

    if "amd64" in found_binaries:
        dst_name = f"{bof_name}.x64.o"
        shutil.copy2(found_binaries["amd64"], dest_dir / dst_name)
        file_entries.append({"os": "windows", "arch": "amd64", "path": dst_name})

    if "386" in found_binaries:
        dst_name = f"{bof_name}.x86.o"
        shutil.copy2(found_binaries["386"], dest_dir / dst_name)
        file_entries.append({"os": "windows", "arch": "386", "path": dst_name})

    return file_entries


def generate_extension_json(cmd_def, file_entries):
    args = []
    for arg_tuple in cmd_def.get("args", []):
        name, arg_type, desc, optional = arg_tuple[0], arg_tuple[1], arg_tuple[2], arg_tuple[3]
        default = arg_tuple[4] if len(arg_tuple) > 4 else None
        entry = {"name": name, "desc": desc, "type": arg_type, "optional": optional}
        if default is not None:
            entry["default"] = default
        args.append(entry)

    if not file_entries:
        file_entries = [
            {"os": "windows", "arch": "amd64", "path": f"{cmd_def['bof']}.x64.o"},
            {"os": "windows", "arch": "386", "path": f"{cmd_def['bof']}.x86.o"},
        ]

    return {
        "command_name": cmd_def["cmd"],
        "help": cmd_def["desc"],
        "extension_author": "Adaptix-Framework",
        "original_author": "Adaptix-Framework",
        "repo_url": "https://github.com/Adaptix-Framework/Extension-Kit",
        "version": "1.0.0",
        "entrypoint": "go",
        "files": file_entries,
        "arguments": args,
    }


def main():
    print("=" * 65)
    print("Extension-Kit (Complete Suite) -> Forge Integration Generator")
    print("=" * 65)

    all_commands = list(BASE_COMMANDS)

    # Dynamic submodules
    submodules = [
        ("AD-BOF/Kerbeus-BOF/kerbeus.axs", "kerbeus", "AD-BOF"),
        ("AD-BOF/ADCS-BOF/ADCS.axs", "certi", "AD-BOF"),
        ("AD-BOF/RelayInformer/RelayInformer.axs", "relay", "AD-BOF"),
        ("AD-BOF/SQL-BOF/SQL.axs", "mssql", "AD-BOF"),
        ("AD-BOF/LDAP-BOF/LDAP.axs", "ldap", "AD-BOF"),
    ]

    for rel_path, prefix, mod in submodules:
        sub_cmds = parse_submodule_axs(rel_path, prefix, mod)
        print(f"  Parsed {len(sub_cmds):2d} commands from {rel_path}")
        all_commands.extend(sub_cmds)

    FORGE_COLLECTIONS_DIR.mkdir(parents=True, exist_ok=True)
    sources = []
    total_commands = 0
    total_with_files = 0

    for cmd_def in all_commands:
        cmd_name = cmd_def["cmd"]
        cmd_dir = FORGE_COLLECTIONS_DIR / cmd_name
        cmd_dir.mkdir(parents=True, exist_ok=True)

        file_entries = copy_and_generate_files(cmd_def, cmd_dir)
        ext_json = generate_extension_json(cmd_def, file_entries)
        ext_path = cmd_dir / "extension.json"
        with open(ext_path, "w") as f:
            json.dump(ext_json, f, indent=2)

        if file_entries:
            total_with_files += 1

        sources.append({
            "name": cmd_name,
            "command_name": cmd_name,
            "description": cmd_def["desc"],
            "repo_url": "",
            "custom_download_url": "",
        })
        total_commands += 1

    sources_path = FORGE_BASE_DIR / "ExtensionKit_sources.json"
    with open(sources_path, "w") as f:
        json.dump(sources, f, indent="\t")
    print(f"\n✓ Wrote {sources_path} ({total_commands} commands)")

    cs_path = FORGE_BASE_DIR / "collection_sources.json"
    with open(cs_path, "r") as f:
        cs = json.load(f)
    if not any(s["name"] == "ExtensionKit" for s in cs):
        cs.append({"name": "ExtensionKit", "type": "bof"})
        with open(cs_path, "w") as f:
            json.dump(cs, f, indent="\t")
        print(f"✓ Added ExtensionKit to {cs_path}")

    cmds_path = FORGE_BASE_DIR / "ExtensionKit_commands.json"
    if not cmds_path.exists():
        with open(cmds_path, "w") as f:
            json.dump([], f)
        print(f"✓ Created empty {cmds_path}")

    print(f"\n{'=' * 65}")
    print(f"Total commands generated: {total_commands}")
    print(f"Commands with .o binaries: {total_with_files}")
    print(f"{'=' * 65}")


if __name__ == "__main__":
    main()
