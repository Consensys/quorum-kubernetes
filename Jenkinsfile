@Library('corda-shared-build-pipeline-steps@5.1') _

pipeline {
    agent {
        docker {
            // Our custom docker image
            image "build-zulu-openjdk:11"
            label "standard"
            registryUrl 'https://engineering-docker.software.r3.com/'
            registryCredentialsId 'artifactory-credentials'
        }
    }
    environment {
        ARTIFACTORY_CREDENTIALS = credentials('artifactory-credentials')
        CORDA_ARTIFACTORY_PASSWORD = "${env.ARTIFACTORY_CREDENTIALS_PSW}"
        CORDA_ARTIFACTORY_USERNAME = "${env.ARTIFACTORY_CREDENTIALS_USR}"
        HELM_CHART_REPO = "helm/charts"
        HELM_REGISTRY = setHelmRegistry()
        HELM_CHARTS_DIR = "helm-charts" // used in the helmPublisher method to set the directory the OCI should be located in.
    }
    options {
        ansiColor('xterm')
        buildDiscarder logRotator(daysToKeepStr: '14')
        disableConcurrentBuilds()
        timestamps()
        timeout(activity: true, time: 10)
    }
    stages {
        stage('Build Helm chart dependencies') {
            steps {
                script {
                    def helmCharts = evaluate("helm/charts")
                    helmCharts.each { helm ->
                        sh "helm dependency build $helm"
                    }
                }
            }
        }
        stage('Publish Helm Chart') {
            steps {
                script {
                    def helmCharts = evaluate("helm/charts")
                    helmCharts.each { helm ->
                        def helmVersion = sh(returnStdout: true, script: "yq '.version' $helm/Chart.yaml").trim()
                        helmPublisher(helm, helmVersion)
                        renderWidget("Published Corda helm chart $helm with version: $helmVersion")
                    }
                }
            }
        }
    }
}

def setHelmRegistry(){
    if(env.TAG_NAME =~ /^release-.*$/) {
         return "engineering-docker-release.software.r3.com"
    } else if (env.BRANCH_NAME =~ /^main$/) {
         return "engineering-docker-release.software.r3.com"
    } else {
        return "engineering-docker-dev.software.r3.com"
    }
}