pipeline {
  triggers { pollSCM('* * * * *') }

  environment {
    imageName = "matthias/myapp"
    dockerImage = ""
  }

  agent { label 'docker' }

  stages {

    stage('Send start notification') {
      steps {
        script {
          slackSend color: '#00aaff', message: "Build #${env.BUILD_NUMBER} started: ${env.JOB_NAME} on branch ${env.BRANCH_NAME}"
        }
      }
    }
    
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
            dockerImage.push('latest')
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

    stage('Remove local images') {
      steps {
        // remove docker images
        sh("docker rmi -f hub.hnrx.de:32769/matthias/myapp:latest || :")
        sh("docker rmi -f hub.hnrx.de:32769/matthias/myapp:$BUILD_NUMBER || :")
        sh("docker rmi -f matthias/myapp:$BUILD_NUMBER || :")
      }
    }
    
    stage('Send success notification') {
      steps {
        script {
          slackSend color: 'good', message: "A new version ${env.BUILD_NUMBER} of " + imageName + " has been deployed. Check it out on https://myapp.hnrx.de"
        }
      }
    }

  }

}
