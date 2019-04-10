

SERVICE_REPO="pskreter/reviewer-service:0.0.6-alpha"
TEST_SERVICE_REPO="pskreter/reviewer-service-test:0.0.8"
REVIEWER_REPO="pskreter/vstsreviewer:1.0.12"

push: build
	docker push ${REVIEWER_REPO}

build:
	docker build -t ${REVIEWER_REPO} -f ./cmd/reviewer/Dockerfile .

build-apiserver:
	docker build -t ${SERVICE_REPO} -f ./cmd/service/Dockerfile . 

push-apiserver: build-apiserver
	docker push ${SERVICE_REPO}


helm-delete:
	helm delete apiserver --purge

run-service: build-service
	docker run -p 8080:8080 ${SERVICE_REPO} --vsts-token ${VSTS_TOKEN} --vsts-username ${VSTS_USERNAME} --mongo-uri ${MONGO_URI} --log-level debug


helm-install:
	helm install --name apiserver ./charts/apiserver --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO}

helm-upgrade:
	helm upgrade --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO} apiserver ./charts/apiserver

###### Test Commands #######
build-test-apiserver:
	docker build -t ${TEST_SERVICE_REPO} -f ./cmd/service/Dockerfile . 

push-test-apiserver: build-test-apiserver
	docker push ${TEST_SERVICE_REPO}

helm-test-install:
	helm install -f ./charts/apiserver/test-values.yaml --name test-apiserver ./charts/apiserver --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${TEST_SERVICE_REPO}

helm-test-upgrade:
	helm upgrade -f ./charts/apiserver/test-values.yaml --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${TEST_SERVICE_REPO} test-apiserver ./charts/apiserver

full-test: push-test-apiserver
	make helm-test-upgrade