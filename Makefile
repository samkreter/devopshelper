



push: build
	docker push ${REVIEWER_REPO}

build:
	docker build -t ${REVIEWER_REPO} -f ./cmd/reviewer/Dockerfile .

build-service:
	docker build -t ${SERVICE_REPO} -f ./cmd/service/Dockerfile . 

run-service: build-service
	docker run -p 8080:8080 ${SERVICE_REPO} --vsts-token ${VSTS_TOKEN} --vsts-username ${VSTS_USERNAME} --mongo-uri ${MONGO_URI} --log-level debug
