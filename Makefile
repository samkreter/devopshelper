

REVIEWER_REPO="pskreter/vstsreviewer:1.0.12"
TEST_SERVICE_REPO="pskreter/reviewer-service-test:0.0.17"

SERVICE_REPO="pskreter/reviewer-service:0.0.21-alpha"
FRONTEND_REPO="pskreter/reviewer-frontend:0.0.7"


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

apiserver-run: build-apiserver
	docker run -p 8080:8080 ${SERVICE_REPO} --vsts-token ${VSTS_TOKEN} --vsts-username ${VSTS_USERNAME} --mongo-uri ${MONGO_URI} --log-level debug


apiserver-deploy:
	helm install --name apiserver ./charts/apiserver --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO}

apiserver-upgrade:
	helm upgrade --set apiserver.token=${VSTS_TOKEN} \
		--reuse-values \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${SERVICE_REPO} apiserver ./charts/apiserver

###### Test Commands #######
test-apiserver-build:
	docker build -t ${TEST_SERVICE_REPO} -f ./cmd/service/Dockerfile . 

test-apiserver-push: test-apiserver-build
	docker push ${TEST_SERVICE_REPO}

apiserver-test-deploy:
	helm install -f ./charts/apiserver/test-values.yaml --name test-apiserver ./charts/apiserver --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${TEST_SERVICE_REPO}

apiserver-test-upgrade:
	helm upgrade -f ./charts/apiserver/test-values.yaml --set apiserver.token=${VSTS_TOKEN} \
		--set apiserver.username=${VSTS_USERNAME} --set apiserver.mongouri=${MONGO_URI} \
		--set apiserver.image=${TEST_SERVICE_REPO} test-apiserver ./charts/apiserver


#### Reviewer V1 #######
push: build
	docker push ${REVIEWER_REPO}

build:
	docker build -t ${REVIEWER_REPO} -f ./cmd/reviewer/Dockerfile .