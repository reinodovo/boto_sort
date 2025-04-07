build: FORCE
	docker build --tag reinodovo/boto_sort . -f build/Dockerfile --platform linux/amd64

push: build
	docker push reinodovo/boto_sort

deploy: FORCE
	-kubectl delete secret boto-sort-secret
	kubectl create secret generic boto-sort-secret --from-env-file=.env
	-kubectl delete -f ./deploy
	kubectl apply -f ./deploy

FORCE: ;