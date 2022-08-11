# vgo - Go version manager

## How to setup and use vgo?

```
# install vgo from source
$ go install github.com/abvarun226/vgo

# run this to enable command completion in shell
$ COMP_INSTALL=1 COMP_YES=1 vgo
installing..

# For zsh shell, open new terminal OR run below command
$ source ~/.zshrc

# For bash shell, open new terminal OR run below command
$ source ~/.bashrc

# download a new version (1.19) for M1 mac (darwin/arm64)
$ vgo download -version 1.19 -platform darwin -arch arm64

# set the active version of go to 1.19
$ vgo set 1.19
go version go1.19 darwin/arm64

# check if 1.19 is activated
$ go version
go version go1.19 darwin/arm64

# download another version (1.18.5) for M1 mac (darwin/arm64)
$ vgo download -version 1.18.5 -platform darwin -arch arm64

# list the installed versions
$ vgo list
Versions: 1.18.5, 1.19
```
