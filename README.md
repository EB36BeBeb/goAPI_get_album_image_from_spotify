# API for AWS to get album cover image from spotify

This API is used for the `spotifycover.ebeb.be` which is to build `N x N` sized grid image using album images of the track.

Supports using `playlist` and `tracks`


## deploy

- set up `token.yml` file with your client token from spotify.
- check `serverless.yml` for the details including `org` `service` `region` etc...
- `make deploy stage=dev` to deploy `dev` 