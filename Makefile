build: FORCE
	docker build --tag reinodovo/boto_sort . -f build/Dockerfile --platform linux/amd64

push: build
	docker push reinodovo/boto_sort

FORCE: ;
