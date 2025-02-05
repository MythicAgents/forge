# Forge

<p align="center">
  <img alt="Forge Logo" src="agent_icons/forge.svg" height="30%" width="30%">
</p>

Forge is a "Command Augmentation" Payload Type for Mythic that brings in @Flangvik's [SharpCollection](https://github.com/Flangvik/SharpCollection) and Sliver's [Armory](https://github.com/sliverarmory/armory/blob/master/armory.json).
Forge itself doesn't build and isn't deployed to a target host; instead, it provides "alias" type commands for BOF execution and .NET execution that are then passed to another Payload Type's commands.
For example, a `forge_net_Rubeus` command might get passed to Apollo's `execute_assembly` command for execution.

Forge comes preconfigured for `Apollo`, but new agent support can always be added on the fly with the `forge_support` command (only needs to happen once per Mythic instance) or by modifying the `payload_type_support.json` file on disk and restarting the container.

## Forge's Icon

forge's icon made by Iconic Panda from [flaticon.com](https://www.flaticon.com/free-icons/forge).

Special shoutout to Alex DeMine for coming up with the name and logo!

## How to install an agent in this format within Mythic

When it's time for you to test out your install or for another user to install your agent, it's pretty simple. Within Mythic you can run the `mythic-cli` binary to install this in one of three ways:

* `sudo ./mythic-cli install github https://github.com/user/repo` to install the main branch
* `sudo ./mythic-cli install github https://github.com/user/repo branchname` to install a specific branch of that repo
* `sudo ./mythic-cli install folder /path/to/local/folder/cloned/from/github` to install from an already cloned down version of an agent repo

Now, you might be wondering _when_ should you or a user do this to properly add your agent to their Mythic instance. There's no wrong answer here, just depends on your preference. The three options are:

* Mythic is already up and going, then you can run the install script and just direct that agent's containers to start (i.e. `sudo ./mythic-cli start agentName` and if that agent has its own special C2 containers, you'll need to start them too via `sudo ./mythic-cli start c2profileName`).
* Mythic is already up and going, but you want to minimize your steps, you can just install the agent and run `sudo ./mythic-cli start`. That script will first _stop_ all of your containers, then start everything back up again. This will also bring in the new agent you just installed.
* Mythic isn't running, you can install the script and just run `sudo ./mythic-cli start`.

## Documentation

The Poseidon documentation source code can be found in the `documenation-payload/forge` directory.
View the rendered documentation by clicking on **Docs -> Agent Documentation** in the upper right-hand corner of the Mythic
interface.

## Building Outside of Mythic

If you want to build outside of Mythic, you can use the `Makefile` included in the project's `Payload_Type/forge` directory. You will need to modify the variables at the top of the Makefile to match the information for connecting to Mythic. 
