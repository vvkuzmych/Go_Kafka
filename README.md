install asdf

Download asdf
Official Download
shell
`git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.1`
Community Supported Download Methods
We highly recommend using the official git method.

Method	    Command
Homebrew	`brew install asdf`
Pacman	    `git clone https://aur.archlinux.org/asdf-vm.git && cd asdf-vm && makepkg -si` or use your preferred AUR helper


Add asdf.sh to your ~/.bashrc with:

shell
`echo -e "\n. \"$(brew --prefix asdf)/libexec/asdf.sh\"" >> ~/.bashrc`
Completions will need to be configured as per Homebrew's instructions or with the following:

shell
`echo -e "\n. \"$(brew --prefix asdf)/etc/bash_completion.d/asdf.bash\"" >> ~/.bashrc
`

after that use ``install-tool-versions.sh`` to install golang  - command in terminal in root of app: `./install-tool-versions.sh`

after successful installation run `docker-compose build`. 
If building was successful then run `docker-compose up` 

Go to Docker and you will see the results of running images. If smth. is not running, try to do it manually in Docker Desktop. 
If all images are running, you should look at logs and see messages that were sent and received. 
