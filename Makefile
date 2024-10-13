build:
	env GOOS=linux go build -ldflags="-s -w" -o getImage/bootstrap getImage/main.go
	env GOOS=linux go build -ldflags="-s -w" -o getPlaylistInfo/bootstrap getPlaylistInfo/main.go
	zip -j getImage/getImage.zip getImage/bootstrap
	zip -j getPlaylistInfo/getPlaylistInfo.zip getPlaylistInfo/bootstrap
    
deploy: build
	AWS_PROFILE=dev serverless deploy --stage $(stage)
clean:
	rm -rf ./bin ./vendor Gopkg.lock ./serverless