def label = UUID.randomUUID().toString()

  podTemplate(label: label, containers: [
    containerTemplate(name: 'go', image: 'soloio/squash-build-container', ttyEnabled: true, command: 'cat',
        resourceRequestCpu: '100m',
        resourceLimitMemory: '1200Mi'),
        containerTemplate(name: 'docker', image: 'docker:17.11', ttyEnabled: true, command: 'cat')
    ], envVars: [
        envVar(key: 'BRANCH_NAME', value: env.BRANCH_NAME),
        envVar(key: 'DOCKER_CONFIG', value: '/etc/docker'),
    ],
    volumes: [hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock'),
              secretVolume(secretName: 'soloio-docker-hub', mountPath: '/etc/docker'),],
    ) {

    node(label) {
      stage('Checkout') {
        checkout scm
        // git 'https://github.com/solo-io/squash'
      }
      stage('Setup go path') {
        container('go') {
          sh 'mkdir -p /go/src/github.com/solo-io/'
          sh 'ln -s $PWD /go/src/github.com/solo-io/squash'
        }
      }
      // TODO: add the go dep's cache as a persistent volume to save time.
      stage('Vendor dependencies') {
        container('go') {
          sh 'cd /go/src/github.com/solo-io/squash/;GO111MODULE=on go mod download'
        }
      }
      stage('Build squash binaries') {
        container('go') {
            sh 'cd /go/src/github.com/solo-io/squash/;make release-binaries'
        }
      }
      stage('Build squash containers') {
        container('docker') {
            sh 'apk add --update git make'
            sh 'make containers'
        }
      }
      stage('Build squash manifests') {
        container('go') {
            sh 'make deployment'
        }
      }
      stage('Push to container registery') {
        container('docker') {
            sh 'make dist'
        }
      }
      /*
      stage('Run e2e test') {
        container('go') {
            // setup a (mini)kube cluster?!
        }
      }
      */
      stage('Archive artifacts') {
        archiveArtifacts 'target/kubernetes/*.yml,target/squash-linux,target/squash-osx,target/squash-windows'
      }
    }
  }
