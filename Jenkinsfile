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
        stage('Check Helm Changes') {
            steps {
                script {
                    def charts = sh(returnStdout: true, script: "ls -d ${env.HELM_CHART_REPO}/*/").trim().split("\n")
                    def newCharts = ""
                        charts.eachWithIndex { chart, i ->
                            if (i == 0) {
                                newCharts = "'$chart':true"
                            }
                            else {
                                newCharts = "$newCharts, '$chart':true"
                            }
                        }
                    echo "Charts configMap created: [${newCharts}]"
                    env.charts_repos = "[$newCharts]"
                }
            }
        }
        stage('Build Helm chart dependencies') {
            steps {
                script {
                    def helmCharts = evaluate(env.charts_repos)
                    helmCharts.each { helm, status ->
                        sh "helm dependency build $helm"
                    }
                }
            }
        }
        stage('Publish Helm Chart') {
            steps {
                script {
                    def helmCharts = evaluate(env.charts_repos)
                    helmCharts.each { helm, status ->
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