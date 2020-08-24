

REVIEWER_REPO="pskreter/vstsreviewer:1.0.12"
TEST_SERVICE_REPO="pskreter/reviewer-service-test:0.0.17"

SERVICE_REPO="pskreter/reviewer-service:0.0.21-alpha"
FRONTEND_REPO="pskreter/reviewer-frontend:0.0.9"


deploy:
	make frontend-deploy
	make apiserver-deploy

upgrade:
	make frontend-upgrade
	make apiserver-upgrade

###### Front End #######
frontend-build:
	docker build -t ${FRONTEND_REPO} -f ./cmd/frontend/Dockerfile ./cmd/frontend

frontend-push: frontend-build
	docker push ${FRONTEND_REPO}

frontend-purge:
	helm delete frontend --purge

frontend-run: build-frontend
	docker run -p 8080:8080 -e PORT=8080 ${FRONTEND_REPO}

frontend-deploy:
	helm install --name frontend ./charts/frontend --set frontend.image=${FRONTEND_REPO}

frontend-upgrade:
	helm upgrade --set frontend.image=${FRONTEND_REPO} frontend ./charts/frontend

###### API Server #######
apiserver-build:
	docker build -t ${SERVICE_REPO} -f ./cmd/service/Dockerfile . 

apiserver-push: apiserver-build
	docker push ${SERVICE_REPO}

apiserver-purge:
	helm delete apiserver --purge

apiserver-run:
	go build -o ./bin/server ./cmd/service
	./bin/server --pat-token ${PAT_TOKEN} --mongo-uri ${MONGO_URI} --log-level debug --mongo-repo-collection=prodRepo

apiserver-deploy:
	helm install --name apiserver ./charts/apiserver --set apiserver.token=${PAT_TOKEN} \
		--set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO}

apiserver-upgrade:
	helm upgrade --set apiserver.token=${PAT_TOKEN} \
		--reuse-values \
	   --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO} apiserver ./charts/apiserver



