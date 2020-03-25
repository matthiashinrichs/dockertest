pipeline {

  environment {
    imageName = "matthias/myapp"
    dockerImage = ""
  }

  agent any

  stages {

    stage('Checkout Source') {
      steps {
        git poll: true,
            credentialsId: 'bitbucket-matthias_hinrichs',
            url: 'https://bitbucket.org/hinrichsmedien/dockertest.git'
      }
    }

    stage('Build image') {
      steps{
        script {
          dockerImage = docker.build imageName + ":$BUILD_NUMBER"
        }
      }
    }

    stage('Push Image') {
      steps{
        script {
          docker.withRegistry( "http://hub.hnrx.de:32769" ) {
            dockerImage.push()
          }
        }
      }
    }  

    stage('Deploy App') {
      steps {
        script {
          kubernetesDeploy(
            configs: "myapp.yaml",
            enableConfigSubstitution: true,
            kubeconfigId: "mykubeconfig")
        }
      }
    }

  }

}
