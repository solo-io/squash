def label = UUID.randomUUID().toString()

  podTemplate(label: label, containers: [
    containerTemplate(name: 'go', image: 'golang:1.9.2-stretch', ttyEnabled: true, command: 'cat',
        resourceRequestCpu: '100m',
        resourceLimitMemory: '1200Mi')
    ], envVars: [
        envVar(key: 'BRANCH_NAME', value: env.BRANCH_NAME)
    ],
    ) {

    node(label) {
      stage('Checkout') {
        // checkout scmsoloio/squash-builder
        git 'https://github.com/solo-io/squash'
      }
      stage('Setup go path') {
        container('go') {
            
          sh 'mkdir -p /go/src/github.com/solo-io/'
          sh 'ln -s $PWD /go/src/github.com/solo-io/squash'
        }
      }
      stage('Vendor dependencies') {
        container('go') {
          sh 'go get -u github.com/golang/dep/cmd/dep'
          sh 'cd /go/src/github.com/solo-io/squash/;dep ensure'          
        }
      }
      stage('Build squash server') {
        container('go') {
            sh 'cd /go/src/github.com/solo-io/squash/;make target/squash-server/squash-server'
        }
      }
    }
  }
