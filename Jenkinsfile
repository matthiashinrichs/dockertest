pipeline {


environment {
  registry = "hub.hnrx.de:32769/"
  dockerImage = "myapp"
}

agent any

stages {

  stage('Checkout Source') {
    steps {
      git credentialsId: 'cbb84c4a-82cd-4b22-8cd7-a8a482711190',
          url: 'https://bitbucket.org/hinrichsmedien/dockertest.git'
    }
  }


stage('Build image') {
steps{
script {
dockerImage = docker.build registry + dockerImage":$BUILD_NUMBER"
}
}
}

}

}