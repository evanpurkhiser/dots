# Dots - A dotfile Management Tool

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

 * Keeping `$HOME` under git version control - This can be difficult and
   dangerous as every directory under your `$HOME` now appears to be in version
   control, so running a `git` command will never fail.

 * Keeping the `$XDG_CONFIG_HOME` directory under version control and using
   environment variables and scripts to force programs to read their
   configurations from the `$XDG_CONFIG_HOME` directory - A great overall
   strategy that keeps dotfiles more organized than keeping them in `$HOME`.

 * Keeping dotfiles in a directory under `$HOME` and  writing an 'installation'
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

 * **Configuration groups**  
   When installing your dotfiles onto a new machine, `dots` offers you the
   ability to select a specific 'group' of dotfiles that you would like to have
   installed into that environment. By organizing your dotfiles into logical
   groups (such as 'machine' groups) it's possible to only install the dotfiles
   that are required by that environment.

 * **Cascading file structure**  
   By selecting multiple configuration groups there is the possibility that two
   groups both contain a dotfile with matching names. For example, if two
   configuration groups both contain a `bashrc` file then the `dots` utility
   will automatically merge these two files together. A special syntax is also
   offered to allow for one file to override another or for the cascading files
   to be merged into the files at specific points.

 * **Installation time includes**  
   Some configuration file formats do not support a way to natively 'include'
   other configuration files into them. The `dots` utility allows for files to
   be inserted into a specific configuration file using a special include
   syntax. The included file will also follow the cascading file structure
   previously mentioned.

 * **Follows [XDG Base Directory Standard](http://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html)**  
   The XDG Base Directory Standard specifies that all configuration files should
   be located in the `$XDG_CONFIG_HOME` directory. By default, this is where all
   configuration files and directories will be installed. While this does
   require a little extra work to ensure all programs read their configuration
   files from here it offers a much more organized view of user dotfiles.

 * **Post installation scripts per file**  
   It's possible to include a `.install` script with any specific dotfile. This
   installation script will be executed any time the specific dotfile is
   installed. This is useful for situations where you _need_ to symbolically
   link a configuration file into `$HOME` or if something needs to be done after
   installation (for example executing :BundleInstall for VIM).

These features are expanded on below.
