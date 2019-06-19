
**Requirements**

- [Git](https://git-scm.com/)
- [Golang](https://golang.org/)
- [dep](https://github.com/golang/dep) (dependency management tool for Go)


**Building**

- *Initial setup*
```bash
mkdir -p $GOPATH/src/github.com/solo-io
cd $GOPATH/src/github.com/solo-io
git clone https://github.com/solo-io/squash.git

cd $GOPATH/src/github.com/solo-io/squash
git checkout -b master
```

- *Build local resources* - sufficient if you only change `squashctl`
  - set a `BUILD_ID`, this will be used as an image tag
```bash
dep ensure -v # do this whenever you add a dependency
BUILD_ID=<build_id> make build -B
```

- *Build and push images* - neccessary if you change code that runs in the cluster (either the plank or secure-mode squash pods)
  - update `solo_project.yaml` to reflect image repos that you have write access to
  - set a `BUILD_ID`, this will be used as an image tag
```bash
BUILD_ID=<build_id> make docker-push -B
```

- Build artifacts can be found in the `_output` directory.


**Note**

_The release build process is defined by the [`cloudbuild.yaml`](https://github.com/solo-io/squash/blob/master/cloudbuild.yaml) file._
