build:
	env GOOS=linux go build -ldflags="-s -w" -o bootstrap main.go
deploy: build
	AWS_PROFILE=dev serverless deploy --stage $(stage)
clean:
	rm -rf ./bin ./vendor Gopkg.lock ./serverless