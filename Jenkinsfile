def projectName = 'k8s-alert-controller'
podTemplate(
  label: projectName, containers: [
    containerTemplate(name: 'jnlp', image: env.JNLP_SLAVE_IMAGE, args: '${computer.jnlpmac} ${computer.name}', alwaysPullImage: true),
    containerTemplate(name: 'kube', image: "${env.PRIVATE_REGISTRY}/library/kubectl:v1.7.2", ttyEnabled: true, command: 'cat'),
    containerTemplate(name: 'golang', image: "golang:${env.GOLANG_VER}", ttyEnabled: true, command: 'cat'),
    containerTemplate(name: 'helm', image: 'henryrao/helm:2.3.1', ttyEnabled: true, command: 'cat'),
    containerTemplate(name: 'dind', image: 'docker:stable-dind', privileged: true, ttyEnabled: true, command: 'dockerd', args: '--host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:2375 --storage-driver=vfs')
  ],
  volumes: [
      emptyDirVolume(mountPath: '/var/run', memory: false),
      hostPathVolume(mountPath: "/etc/docker/certs.d/${env.PRIVATE_REGISTRY}/ca.crt", hostPath: "/etc/docker/certs.d/${env.PRIVATE_REGISTRY}/ca.crt"),
      hostPathVolume(mountPath: '/home/jenkins/.kube/config', hostPath: '/etc/kubernetes/admin.conf'),
      persistentVolumeClaim(claimName: env.HELM_REPOSITORY, mountPath: '/var/helm/', readOnly: false)
  ]) {
    node(projectName) {
        ansiColor('xterm') {
            stage('git clone') {
                checkout scm
            }

            def image
            def last_commit = sh(script: 'git log --format=%B -n 1', returnStdout: true).trim()

            def gitBranchName = sh(script: 'git branch -r | cut -d\'/\' -f 2', returnStdout: true).trim()
            def gitComitHash = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
            def imageTag = "${gitBranchName}-${env.BUILD_ID}-${gitComitHash}"
            sh "echo '${imageTag}'"

            stage('build image'){
                sh "docker build -t ${env.PRIVATE_REGISTRY}/library/${projectName}:${imageTag} ."
            }

            stage('push image') {
                withDockerRegistry(url: env.PRIVATE_REGISTRY_URL, credentialsId: 'docker-login') {
                    image = docker.image("${env.PRIVATE_REGISTRY}/library/${projectName}:${imageTag}")
                    image.push()
                    image.push('latest')
                }
            }

            container('helm') {
                sh 'helm init --client-only'

                def releaseName = "${projectName}-release-${env.BUILD_ID}"

                try {
                    sh 'pwd;ls -al'
                    dir("${projectName}") {
                        stage('test chart') {
                            sh 'helm repo add grandsys https://grandsys.github.io/helm-repository/'
                            sh 'helm dependency update'
                            echo 'syntax check'
                            sh 'helm lint .'

                            echo 'install chart'
                            def service = "${projectName}-test-${env.BUILD_ID}"
                            sh "helm install --set=service.name=${service} -n ${releaseName} ."
                            sh "helm test ${releaseName} --cleanup"
                        }
                    }

                    stage('package chart') {
                        dir("${projectName}") {
                            echo 'archive chart'
                            sh 'helm package --destination /var/helm/repo .'
                            
                            echo 'generate an index file'
                            sh """
                            merge=`[[ -e '/var/helm/repo/index.yaml' ]] && echo '--merge /var/helm/repo/index.yaml' || echo ''`
                            helm repo index --url ${env.HELM_PUBLIC_REPO_URL} \$merge /var/helm/repo
                            """
                        }
                        build job: 'helm-repository/master', parameters: [string(name: 'commiter', value: "${env.JOB_NAME}\ncommit: ${last_commit}")]
                    }

                } catch (error) {
                    echo "${e}"
                    currentBuild.result = FAILURE
                } finally {
                    stage('clean up') {
                        container('helm') {
                            sh "helm delete --purge ${releaseName}"
                        }
                        container('kube') {
                            sh "kubectl delete pvc -l release=${releaseName}"
                        }
                    }
                }
            }
        }
    }
}