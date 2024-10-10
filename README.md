# API for AWS to get album cover image from spotify

This API is used for the [spotifycover.ebeb.be](https://spotifycover.ebeb.be) which is to build `N x N` sized grid image using album images of the track.

Supports using `playlist` and `tracks`


## deploy

- set up `token.yml` file with your client token from spotify.
- check `serverless.yml` for the details including `org` `service` `region` etc...
- `make deploy stage=dev` to deploy `dev` 


## Request Example

### Make 5x5 grid using playlist

```
{
  "type": "playlist",
  "size": 5,
  "urls": ["https://open.spotify.com/playlist/0HuPdv4PSfM9FKjTEUFWs4?si=15022b8fbb8f4a19"]
}
```

### Make 2x2 grid Using tracks

```
{
  "type": "playlist",
  "size": 5,
  "urls": ["https://open.spotify.com/track/YOUR_SONG_TRACK_ID?si=15022b8fb","https://open.spotify.com/track/YOUR_SONG_TRACK_ID","https://open.spotify.com/track/YOUR_SONG_TRACK_ID","https://open.spotify.com/track/YOUR_SONG_TRACK_ID"]
}
```
