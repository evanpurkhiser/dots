## ⚠️ NOTE

> I have rewrote this tool in go, you can still find the old python 2.x version
> of this tool on the `releases/1.x` branch. The readme below **is still for
> the 1.x version of this tool!**

# Dots - A dotfile Management Tool

TODO: Add contents section here

In the desktop \*nix world "dotfiles" are the life and blood behind the
customization of your environment. Everything from custom bash prompts to window
manager configurations, dotfiles define your work flow. Because of this, it has
become a very common practice to keep your dotfiles in a git repository and even
make the repository available on GitHub for the community to explore.

### The Problem

Keeping your dotfiles in a git repository can prove to be rather difficult
however. Because dotfiles can be spread all across the users `$HOME` the
process of tracking these files can be tedious. A few strategies have emerged
within the community that offers command over dotfiles:

- Keeping `$HOME` under git version control - This can be difficult and
  dangerous as every directory under your `$HOME` now appears to be in version
  control, so running a `git` command will never fail.

- Keeping the `$XDG_CONFIG_HOME` directory under version control and using
  environment variables and scripts to force programs to read their
  configurations from the `$XDG_CONFIG_HOME` directory - A great overall
  strategy that keeps dotfiles more organized than keeping them in `$HOME`.

- Keeping dotfiles in a directory under `$HOME` and writing an 'installation'
  script to symbolically link the files to their appropriate location in the
  users `$HOME` - While this method is rather robust, it can be difficult to
  manage a installation script specifically for your dot files, and may impose
  limits in how you organize your dotfiles.

However, none of these methods offer any sort of ability to manage dot files
across multiple machines and environments. While you can use any one of these
strategies to effectively manage dotfiles for a single machine, using the same
dotfiles on another machine _will be_ convoluted and cumbersome.

For example, say you manage a set of dotfiles for your desktop machine that
includes configuration for bash, Vim, a window manager, and GTK theme settings.
If you would like to use these dotfiles in a headless server environment then
you would obviously not need your window manager and GTK configurations. Even
worse, changes may need to be made to some of the configuration files for that
specific environment.

### The Solution

The `dots` utility aims to help make the installation, revision control, and
organization of dot files easy and intuitive. The dots utility solves all of the
problems that the solutions mentioned above solve while also facilitating the
ability to maintain a single repository of dotfiles that can be installed on any
machine you access and would like to have your environment configured in a snap.

`dots` offers the following features:

- **Configuration groups**  
  When installing your dotfiles onto a new machine, `dots` offers you the
  ability to select a specific 'group' of dotfiles that you would like to have
  installed into that environment. By organizing your dotfiles into logical
  groups (such as 'machine' groups) it's possible to only install the dotfiles
  that are required by that environment.

- **Cascading file structure**  
  By selecting multiple configuration groups there is the possibility that two
  groups both contain a dotfile with matching names. For example, if two
  configuration groups both contain a `bashrc` file then the `dots` utility
  will automatically merge these two files together. A special syntax is also
  offered to allow for one file to override another or for the cascading files
  to be merged into the files at specific points.

- **Follows [XDG Base Directory Standard](http://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html)**  
  The XDG Base Directory Standard specifies that all configuration files should
  be located in the `$XDG_CONFIG_HOME` directory. By default, this is where all
  configuration files and directories will be installed. While this does
  require a little extra work to ensure all programs read their configuration
  files from here it offers a much more organized view of user dotfiles.

- **Post installation scripts per file**  
  It's possible to include a `.install` script with any specific dotfile or
  directory. This installation script will be executed any time the specific
  dotfile is installed. This is useful for situations where you _need_ to
  symbolically link a configuration file into `$HOME` or if something needs to
  be done after installation (for example executing :BundleInstall for VIM).

These features are expanded on below.

## Usage and configuration

**NOTE:**
For a quick and easy way to get going managing your dot files more effectively
see the [dots-template](https://github.com/evanpurkhiser/dots-template)
repository that offers instructions for getting setup.

---

It's recommended to read through the entirety of this README to have a good
understanding of how the dots management utility works. Here are a few key points
to keep in mind however:

- Configuration files will be installed into `$XDG_CONFIG_HOME`.
- Configuration group directories are to be located in `$HOME/.local/etc`.
- A PKGBUILD file is also available for Arch Linux [on the
  AUR](https://aur.archlinux.org/packages/dots-manager/).
- The `dots` binary should be made available in your `PATH`.
- The `dots` binary should support Python > 2.7 / Python > 3.2.
- See [Evan Purkhisers personal
  dotfiles](https://github.com/evanpurkhiser/dots-personal) for an example
  configuration.

For details on using the `dots` tool itself see the `dots help` [USAGE
output](bin/dots#L83).

### The initialization bootstrap script

A [initialization script](contrib/initialize) is included in the `contrib`
directory, providing an easy way to initalize your dotfiles into the proper
directory, temporarily setup the `PATH` for the dots, and temporarily source the
bash completion scripts. This way you can quickly setup your dotfiles, activate
your configuration groups, and install the dotfiles themselves

For example:

```sh
$ cd ~
$ git clone https://github.com/Your/Dotfiles
$ DOTS_CLONE_DIR=~/Dotfiles
$ source dots/contrib/initialize
```

This will do the following:

1.  Move the Dotfiles into `$HOME/.local/etc`
2.  Symbolically link the dots executable into `$HOME/.local/bin`
3.  Add `$HOME/.local/bin` into the `PATH` if it's not already
4.  Source the `contrib/bash_completion` script

You can then setup your dotfiles using the `dots` command:

```sh
$ dots groups set base machines/desktop
$ dots install
```

### Bash completion

A bash [completion script](contrib/bash_completion) is included and provides
completions for all aspects of the `dots` command. If you would like to take
advantage of this it's recommended that you source this file in your `bashrc`.

## Configuration groups

The primary feature of the dots utility is to allow for dotfiles to be organized
into different "Configuration groups". These configuration groups can then be
enabled or disabled for the specific environment that the dotfiles are being
installed into.

Configuration groups are two-level directories containing configuration files
and directories that will be installed when the group is activated. The dots
utility also includes a special configuration group that is hard coded into the
utility: The `base` group is a single-level directory, so all files and
directories located in the `base` directory will be installed if the base group
is activated.

## Extending and Overriding configurations

Configuration groups can override or extend files that are included in
configuration groups specified prior to them.

### Extending configuration files

If a configuration file in the `base` group specifies _most_ of what is needed,
but for the specific environment you're installing the configuration files into
requires a little extra configuration for that file it is possible to append to
it.

For example, if you would like to add more options to the `bashrc` for your
`machine/desktop` group, you can simply include the `bash/bashrc` file and it
will automagically be appended to the `base/bash/bashrc` file upon installation.

Shebangs will be removed from the first line of the file being appended.

### Overriding

You can completely override a configuration file included in a previous group.
This is similar to extending a configuration file as described in the previous
section, however the file will simply replace the file specified in the
environment group.

Enable overriding for a file by appending `.override` to the filename.

## Installation Scripts

For each configuration file you may also include a `.install` script. This file
will be executed when the specific configuration file has been installed. If the
destination file has not been changed from the compiled file then the install
script will not be executed.

The installation scripts will be executed with the destination directory as the
current working directory. In order for the scripts to be executed, they must
be executable and include a shebang.

For example: We have a `base/vim/vimrc` configuration file. We could also
include a `base/vim/vimrc.install` file that executes some commands when the
vimrc file is installed. The script will be executed with the
`$XDG_CONFIG_HOME/vim/` as the working directory.
