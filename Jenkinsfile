def label = UUID.randomUUID().toString()

timestamps {

  podTemplate(label: label, containers: [
    containerTemplate(name: 'maven', image: 'maven:3.5.0-jdk-8-alpine', ttyEnabled: true, command: 'cat',
        resourceRequestCpu: '100m',
        resourceLimitMemory: '1200Mi')
    ], envVars: [
        envVar(key: '_JAVA_OPTIONS', value: jvmOptions),
        envVar(key: 'BRANCH_NAME', value: env.BRANCH_NAME)
    ],
    ) {

    node(label) {
      stage('Checkout') {
        checkout scm
      }
      stage('Build squash server') {
          sh 'make target/squash-server/squash-server'
      }
    }
  }

}