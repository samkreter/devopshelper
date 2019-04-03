



push: build
	docker push ${REVIEWER_REPO}

build:
	docker build -t ${REVIEWER_REPO} -f ./cmd/reviewer/Dockerfile .


build-service:
	docker build -t ${SERVICE_REPO} -f ./cmd/service/Dockerfile . 
