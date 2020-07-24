.PHONY: FORCE

NAME=combiner

build: FORCE 
	go build -o ./build/${NAME}