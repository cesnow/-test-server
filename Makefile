

deploy-gatesession:
	docker-compose -p web_librechat -f deployments/docker-compose/gatesession/docker-compose.yaml up --build -d