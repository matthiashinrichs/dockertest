pipeline {


environment {
registry = "192.168.1.81:5000/"
dockerImage = ""
}

agent any

stages {

stage('Checkout Source') {
steps {
git 'https://matthias_hinrichs@bitbucket.org/hinrichsmedien/dockertest.git'
}
}


stage('Build image') {
steps{
script {
dockerImage = docker.build registry + ":$BUILD_NUMBER"
}
}
}

}

}