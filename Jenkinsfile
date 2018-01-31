properties([
    buildDiscarder(logRotator(
        artifactDaysToKeepStr: '14',
        artifactNumToKeepStr: '30',
        daysToKeepStr: '14',
        numToKeepStr: '30',
    )),
    disableConcurrentBuilds(),
    pipelineTriggers([]),
    parameters([
        booleanParam(name: 'BUILD_RELEASE', defaultValue: false, description: ''),
        booleanParam(name: 'USE_BRANCH_AS_TAG', defaultValue: false, description: ''),
        booleanParam(name: 'RUN_INTEGRATION_TESTS', defaultValue: true, description: 'If true, run integration tests even if branch is not master'),
        booleanParam(name: 'SHORT_TESTS', defaultValue: false, description: 'If true, run tests with -test.short=true for running a subset of tests'),
        booleanParam(name: 'SKIP_DOCKER_STAGES', defaultValue: false, description: 'If true, skips docker build, tag and push'),
        booleanParam(name: 'SKIP_NAMESPACE_CLEANUP', defaultValue: false, description: 'If true, skips deleting the Kubernetes namespace at the end of the job'),
    ])
])

def isPullRequest = env.BRANCH_NAME.startsWith("PR-")
def isMasterBranch = env.BRANCH_NAME == "master"

def instanceCap = isMasterBranch ? 1 : 5
def podLabel = "kube-chargeback-build-${isMasterBranch ? 'master' : 'pr'}"

def awsBillingBucket = "team-chargeback"
def awsBillingBucketPrefix = "cost-usage-report/team-chargeback-chancez/"

podTemplate(
    cloud: 'kubernetes',
    containers: [
        containerTemplate(
            alwaysPullImage: false,
            envVars: [],
            command: 'dockerd-entrypoint.sh',
            args: '--storage-driver=overlay',
            image: 'docker:dind',
            name: 'docker',
            // resourceRequestCpu: '1750m',
            // resourceRequestMemory: '1500Mi',
            privileged: true,
            ttyEnabled: true,
        ),
    ],
    volumes: [
        emptyDirVolume(
            mountPath: '/var/lib/docker',
            memory: false,
        ),
    ],
    idleMinutes: 5,
    instanceCap: 5,
    label: podLabel,
    name: podLabel,
) {
    node (podLabel) {
    timestamps {
        def runIntegrationTests = isMasterBranch || params.RUN_INTEGRATION_TESTS || (isPullRequest && pullRequest.labels.contains("run-integration-tests"))
        def shortTests = params.SHORT_TESTS || (isPullRequest && pullRequest.labels.contains("run-short-tests"))

        def gopath = "${env.WORKSPACE}/go"
        def kubeChargebackDir = "${gopath}/src/github.com/coreos-inc/kube-chargeback"

        def gitCommit
        def gitTag
        def branchTag = env.BRANCH_NAME.toLowerCase()
        def deployTag = "${branchTag}-${currentBuild.number}"
        def chargebackNamespace = "chargeback-ci-${branchTag}"

        try {
            container('docker'){
                stage('checkout') {
                    sh """
                    apk update
                    apk add git bash jq zip
                    """

                    checkout([
                        $class: 'GitSCM',
                        branches: scm.branches,
                        extensions: scm.extensions + [[$class: 'RelativeTargetDirectory', relativeTargetDir: kubeChargebackDir]],
                        userRemoteConfigs: scm.userRemoteConfigs
                    ])

                    gitCommit = sh(returnStdout: true, script: "cd ${kubeChargebackDir} && git rev-parse HEAD").trim()
                    gitTag = sh(returnStdout: true, script: "cd ${kubeChargebackDir} && git describe --tags --exact-match HEAD 2>/dev/null || true").trim()
                    echo "Git Commit: ${gitCommit}"
                    if (gitTag) {
                        echo "This commit has a matching git Tag: ${gitTag}"
                    }

                    if (params.BUILD_RELEASE) {
                        if (params.USE_BRANCH_AS_TAG) {
                            gitTag = branchTag
                        } else if (!gitTag) {
                            error "Unable to detect git tag"
                        }
                        deployTag = gitTag
                    }
                }
            }

            withCredentials([
                [$class: 'FileBinding', credentialsId: 'chargeback-ci-kubeconfig', variable: 'KUBECONFIG'],
                [$class: 'AmazonWebServicesCredentialsBinding', credentialsId: 'kube-chargeback-s3', accessKeyVariable: 'AWS_ACCESS_KEY_ID', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'],
                usernamePassword(credentialsId: 'quay-coreos-jenkins-push', passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME'),
            ]) {
                withEnv([
                    "GOPATH=${gopath}",
                    "USE_LATEST_TAG=${isMasterBranch}",
                    "BRANCH_TAG=${branchTag}",
                    "DEPLOY_TAG=${deployTag}",
                    "CHARGEBACK_NAMESPACE=${chargebackNamespace}",
                    "CHARGEBACK_SHORT_TESTS=${shortTests}",
                    "KUBECONFIG=${KUBECONFIG}",
                    "CHARGEBACK_NAMESPACE=${chargebackNamespace}",
                    "ENABLE_AWS_BILLING=true",
                    "AWS_BILLING_BUCKET=${awsBillingBucket}",
                    "AWS_BILLING_BUCKET_PREFIX=${awsBillingBucketPrefix}",
                    "AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}",
                    "AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}",
                ]){
                    container('docker'){
                        echo "Authenticating to docker registry"
                        sh 'docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD quay.io'

                        stage('install dependencies') {
                            // Build & install thrift
                            sh '''#!/bin/bash
                            set -e
                            apk add make go libc-dev curl
                            export HELM_VERSION=2.8.0
                            curl \
                                --silent \
                                --show-error \
                                --location \
                                "https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz" \
                                | tar xz --strip-components=1 -C /usr/local/bin linux-amd64/helm \
                                && chmod +x /usr/local/bin/helm
                            helm init --client-only --skip-refresh
                            helm repo remove stable || true

                            export KUBERNETES_VERSION=1.8.3
                            curl \
                                --silent \
                                --show-error \
                                --location \
                                "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" \
                                -o /usr/local/bin/kubectl \
                                 && chmod +x /usr/local/bin/kubectl
                            '''
                        }

                        dir(kubeChargebackDir) {
                            stage('test') {
                                sh """#!/bin/bash
                                make k8s-verify-codegen
                                """
                            }

                            stage('build') {
                                if (params.SKIP_DOCKER_STAGES) {
                                    echo "Skipping docker build"
                                } else if (!params.BUILD_RELEASE) {
                                    ansiColor('xterm') {
                                        sh """#!/bin/bash -ex
                                        make docker-build-all -j 2 \
                                            USE_LATEST_TAG=${USE_LATEST_TAG} \
                                            BRANCH_TAG=${BRANCH_TAG}
                                        """
                                    }
                                } else {
                                    // Images should already have been built if
                                    // we're doing a release build. In the tag
                                    // stage we will pull and tag these images
                                    echo "Release build, skipping building of images."
                                }
                            }

                            stage('tag') {
                                if (params.SKIP_DOCKER_STAGES) {
                                    echo "Skipping docker tag"
                                } else if (!params.BUILD_RELEASE) {
                                    ansiColor('xterm') {
                                        sh """#!/bin/bash -ex
                                        make docker-tag-all -j 2 \
                                            IMAGE_TAG=${DEPLOY_TAG}
                                        """
                                    }
                                } else {
                                    ansiColor('xterm') {
                                        sh """#!/bin/bash -ex
                                        make docker-tag-all \
                                            PULL_TAG_IMAGE_SOURCE=true \
                                            IMAGE_TAG=${gitTag}
                                        """
                                    }
                                }
                            }

                            stage('push') {
                                if (params.SKIP_DOCKER_STAGES) {
                                    echo "Skipping docker push"
                                } else if (!params.BUILD_RELEASE) {
                                    sh """#!/bin/bash -ex
                                    make docker-push-all -j 2 \
                                        USE_LATEST_TAG=${USE_LATEST_TAG} \
                                        BRANCH_TAG=${BRANCH_TAG}
                                    # Unset BRANCH_TAG so we don't push the same
                                    # image twice
                                    unset BRANCH_TAG
                                    make docker-push-all -j 2 \
                                        IMAGE_TAG=${DEPLOY_TAG}
                                        BRANCH_TAG=
                                    """
                                } else {
                                    sh """#!/bin/bash -ex
                                    make docker-push-all -j 2 \
                                        USE_LATEST_TAG=false \
                                        IMAGE_TAG=${gitTag}
                                    """
                                }
                            }

                            stage('release') {
                                if (params.BUILD_RELEASE) {
                                    sh """#!/bin/bash -ex
                                    make release RELEASE_VERSION=${BRANCH_TAG}
                                    """
                                    archiveArtifacts artifacts: 'tectonic-chargeback-*.zip', fingerprint: true, onlyIfSuccessful: true
                                } else {
                                    echo "Skipping release step, not a release"
                                }
                            }

                            stage('deploy') {
                                if (runIntegrationTests ) {
                                    echo "Deploying chargeback"

                                    ansiColor('xterm') {
                                        timeout(10) {
                                            sh """#!/bin/bash
                                            ./hack/deploy-ci.sh
                                            """
                                        }
                                    }
                                    echo "Successfully deployed chargeback-helm-operator"
                                } else {
                                    echo "Non-master branch, skipping deploy"
                                }
                            }
                            stage('integration tests') {
                                if (runIntegrationTests) {
                                    echo "Running chargeback integration tests"

                                    ansiColor('xterm') {
                                        sh """#!/bin/bash
                                        ./hack/integration-tests.sh
                                        """
                                    }
                                } else {
                                    echo "Non-master branch, skipping chargeback integration test"
                                }
                            }
                        }
                    }
                }
            }
        } catch (e) {
            // If there was an exception thrown, the build failed
            echo "Build failed"
            currentBuild.result = "FAILED"
            throw e
        } finally {
            if (runIntegrationTests && !params.SKIP_NAMESPACE_CLEANUP) {
                withCredentials([
                    [$class: 'FileBinding', credentialsId: 'chargeback-ci-kubeconfig', variable: 'KUBECONFIG'],
                ]) {
                    withEnv([
                        "CHARGEBACK_NAMESPACE=${chargebackNamespace}",
                    ]){
                        container("docker") {
                            dir(kubeChargebackDir) {
                                sh '''#!/bin/bash
                                source hack/util.sh
                                export KUBECONFIG=${KUBECONFIG}
                                CHARGEBACK_NAMESPACE="$(sanetize_namespace "$CHARGEBACK_NAMESPACE")"
                                kubectl delete ns --now $CHARGEBACK_NAMESPACE
                                '''
                            }
                        }
                    }
                }
            }
            cleanWs notFailBuild: true
            // notifyBuild(currentBuild.result)
        }
    }
} // timestamps end
} // podTemplate end
