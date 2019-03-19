



push: build
	docker push ${REVIEWER_REPO}

build:
	docker build -t ${REVIEWER_REPO} -f ./cmd/reviewer/Dockerfile .
