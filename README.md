# vmpooler-bitbar

## What is this?

Vmpooler-bitbar is a small ruby plugin for [@matryer's BitBar application](https://github.com/matryer/bitbar) built on top of [@braincain's vmfloaty](https://github.com/briancain/vmfloaty) which shows the status of all of your vmpooler instances and allows quick access to actions such as ssh'ing to a node or deleting an instance... and more.

Too much talk, have a look at it in action.

![demo video showing vmpooler-bitbar in action](https://raw.githubusercontent.com/johnmccabe/vmpooler-bitbar/gh-pages/images/vmpooler-bitbar.gif)

## Features

- updates every `30s`, see [here](https://github.com/matryer/bitbar#configure-the-refresh-time) if you want to change the refresh interval - I don't recommend setting it lower than 30s.
- shows all active vms created using your token
- vms with < 1hr before their deletion are highlighted in red
- quick access to some details of each vm, tags etc
- ssh directly to a vm from the menu
  - OSX Terminal supported by default
  - [iTerm2 can be used instead](#using-iterm2-instead-of-osx-terminal)
- delete a vm from the menu
- extend the lifetime of a vm from the menu
- delete all vms from the menu
- extend the lifetime of all vms from the menu
- [click on an item to copy it to the clipboard](#copying-hostname-etc)
- create a new vm from the menu (available templates pulled from vmpooler, with new vms tagged with `created_by=vmpooler-bitbar`)
- integrates with the OSX Notification Centre

## Getting Started

### Prerequisites

- vmfloaty should be installed and configured, with the vmpooler `url`, `user` and `token` set in your `~/.vmfloaty.yml` config file (see the [vmfloaty docs](https://github.com/briancain/vmfloaty#example-workflow) for information on obtaining a token). If you are able to run `floaty token status` then you should be good to go.
- for the SSH to vmpooler instance action to work you should have the vmpooler ssh key added to the ssh agent, `ssh-add /path/to/priv/key`.

### Install BitBar

If you don't already have BitBar installed you can install using `brew` or by grabbing a release directly from [GitHub](https://github.com/matryer/bitbar/releases/tag/v1.9.1). If you already have BitBar installed you can jump to installing and running the plugin.

    brew cask install bitbar

You can now start BitBar from the `Applications` folder or:

    open /Applications/BitBar.app

If this is your first time installing BitBar you will be prompted to choose/create a plugins directory, for example `~/Documents/bitbar_plugins/`.

Any executable scripts copied to this directory will be rendered in the menubar by BitBar and it is here we will copy the vmpooler-bitbar script.

### Add the vmpooler-bitbar plugin

Copy `vmpooler-bitbar.30s.rb` to your BitBar plugins directory.

From the BitBar menu select `refresh all` to have BitBar rescan the plugins directory and you should see the `VM: <number of vms>` appear in your menubar.

## Tips

### Copying Hostname etc
To copy displayed text to the clipboard just click on the item in the menu, this is currently supported for:

- VM hostname (note that the full fqdn will be copied)
- Any Status or Tag entries in the VM submenu

### Using iTerm2 instead of OSX Terminal
To use iTerm2 you must first configure it as described in the _'How do I set iTerm2 as the handler for ssh:// links'_ section of the [iTerm2 FAQ page](http://iterm2.com/faq.html).

If after making the changes above the OSX Terminal continues to open then you should run the following command to rebuild the launch services DB (via [iTerm2 issue #5022](https://gitlab.com/gnachman/iterm2/issues/5022))

    /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -kill -r -domain local -domain system -domain user

## Troubleshooting

### Errors throwing during brew install
If you encounter the following error:

    Error: Cask 'bitbar' definition is invalid: Bad header line: parse failed

You will need to fix your brew cask before reattempting to install BitBar.

    brew uninstall --force brew-cask; brew update
