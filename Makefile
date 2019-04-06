



push: build
	docker push ${REVIEWER_REPO}

build:
	docker build -t ${REVIEWER_REPO} -f ./cmd/reviewer/Dockerfile .

build-apiserver:
	docker build -t ${SERVICE_REPO} -f ./cmd/service/Dockerfile . 

push-apiserver: build-apiserver
	docker push ${SERVICE_REPO}

helm-install:
	helm install --name apiserver ./charts/apiserver --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO}

helm-upgrade:
	helm upgrade --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO} apiserver ./charts/apiserver

helm-delete:
	helm delete apiserver --purge

run-service: build-service
	docker run -p 8080:8080 ${SERVICE_REPO} --vsts-token ${VSTS_TOKEN} --vsts-username ${VSTS_USERNAME} --mongo-uri ${MONGO_URI} --log-level debug
