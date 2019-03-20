## How to Build Squash

### Pre-requirement
* [Git](https://git-scm.com/)
  * MAC(need [homebrew](https://docs.brew.sh/Installation))
    ```bash
      brew install git
    ```

  * Ubuntu
    ```bash
      sudo apt install git
    ```

* golang
  * [Manual install](https://golang.org/dl/)
  * MAC(need [homebrew](https://docs.brew.sh/Installation))
    ```bash
      brew install go
    ```
  * Ubuntu
    ```bash
      sudo apt update
      sudo apt install golang-go
    ```

* [dep](https://github.com/golang/dep) (dependency management tool for Go)
  * MAC(need [homebrew](https://docs.brew.sh/Installation))
    ```bash
      brew install dep
      brew upgrade dep
    ```

  * Ubuntu
    ```bash
      sudo apt install go-dep
    ```

### Setting up Go Environment

* MAC
  ```bash
    # Set env variables and add Go to the PATH in .bashrc / .zshrc file.
    export GOPATH=$HOME/golang
    export GOROOT=/usr/local/opt/go/libexec
    export PATH=$PATH:$GOPATH/bin:$GOROOT/bin
  ```

* Ubuntu
  ```bash
    # Set env variables and add Go to the PATH in .bashrc / .zshrc file.
    export GOPATH=$HOME/golang
    export GOROOT=/usr/lib/go
    export PATH=$PATH:$GOPATH/bin:$GOROOT/bin
  ```

* make $GOPATH directory
  ```bash
    mkdir -p $GOAPTH/src
  ```

### Download the Squash source and install the dependencies

* download the Squash source
  ```bash
    # Download the source by using go get
    go get github.com/solo-io/squash
  
    # or by using git clone
    # mkdir -p $GOPATH/src/github.com/solo-io
    # cd $GOPATH/src/github.com/solo-io
    # git clone https://github.com/solo-io/squash.git
    
    cd $GOPATH/src/github.com/solo-io/squash
    git checkout -b v0.5.5 tags/v0.5.5
  ```

* install the dependencies
  ```bash
    # This job may take some time.
    # To see the progress, check the pkg directory size:
    # cd $GOPATH/pkg ; du -sh
    dep ensure
  ```

### Now you can build Squash
  ```bash
    make build
  ```

