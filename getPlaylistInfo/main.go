package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type RequestBody struct {
	Url string `json:"url"`
}

type TrackInfo struct {
	TrackName        string `json:"trackName"`
	ArtistName       string `json:"artistName"`
	AlbumName        string `json:"albumName"`
	AlbumReleaseDate string `json:"albumReleaseDate"`
}

type ResponseBody struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Info    []TrackInfo `json:"info"`
}

func auth() string {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	authURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Add("Authorization", "Basic "+authHeader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Fatal(err)
		}

		token := result["access_token"].(string)
		fmt.Println("Access token:", token)
		return token
	} else {
		fmt.Printf("Failed to get access token: %d\n", resp.StatusCode)
	}
	return ""
}

func extractSpotifyObjectID(url string) string {
	// Check if "spotify" is in the URL
	if !strings.Contains(url, "spotify") {
		return url
	}

	// Remove query parameters by splitting at "?si"
	url = strings.Split(url, "?si")[0]

	// Define the regular expression pattern
	pattern := fmt.Sprintf(`https://open\.spotify\.com/%s/([a-zA-Z0-9]+)`, "playlist")

	// Compile the regular expression
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return url
	}

	// Find the match using the regular expression
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1] // Return the first capture group
	}

	// If no match is found, return the original URL
	return url
}

func getPlaylistItemInfo(id string, accessToken string) ([]TrackInfo, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/", id) + "tracks?fields=items%28track%28name%2Calbum%28name%2Crelease_date%29%2Cartists%28name%29%29"
	fmt.Println(url)
	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Authorization header
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200)
	if resp.StatusCode == 200 {
		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Parse the JSON response
		var result map[string]interface{}

		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%v", ExtractPlaylistInfo(result))

		return ExtractPlaylistInfo(result), nil

	} else {
		// Print error message if request was not successful
		fmt.Printf("Error: %v\n", resp.Status)
	}
	return nil, err
}

func ExtractPlaylistInfo(data map[string]interface{}) []TrackInfo {
	var trackInfos []TrackInfo

	// items를 순회하며 TrackInfo로 변환
	if items, ok := data["items"].([]interface{}); ok {
		for _, item := range items {
			if track, ok := item.(map[string]interface{})["track"].(map[string]interface{}); ok {
				album := track["album"].(map[string]interface{})
				artists := track["artists"].([]interface{})
				artistNames := []string{}

				for _, artist := range artists {
					if artistMap, ok := artist.(map[string]interface{}); ok {
						artistNames = append(artistNames, artistMap["name"].(string))
					}
				}

				// 아티스트 이름들을 쉼표로 구분된 문자열로 합침
				artistName := ""
				if len(artistNames) > 0 {
					artistName = joinStrings(artistNames, ", ")
				}

				trackInfo := TrackInfo{
					TrackName:        track["name"].(string),
					ArtistName:       artistName,
					AlbumName:        album["name"].(string),
					AlbumReleaseDate: album["release_date"].(string),
				}

				trackInfos = append(trackInfos, trackInfo)
			}
		}
	}
	return trackInfos
	// // 결과 출력
	// result, _ := json.MarshalIndent(trackInfos, "", "  ")
	// fmt.Println(string(result))
}

// joinStrings 함수: 슬라이스를 문자열로 합치는 함수
func joinStrings(elements []string, separator string) string {
	if len(elements) == 0 {
		return ""
	}
	result := ""
	for i, element := range elements {
		if i > 0 {
			result += separator
		}
		result += element
	}
	return result
}

func buildResponse(status int, msg string, info []TrackInfo) ResponseBody {
	return ResponseBody{
		Status:  status,
		Message: msg,
		Info:    info,
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (ResponseBody, error) {

	var body RequestBody
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		log.Fatal(err)
	}
	url := body.Url
	singleId := extractSpotifyObjectID(url)
	fmt.Println(singleId)
	if singleId == url {
		return buildResponse(400, "Invalid Url included : ", nil), nil
	}

	accessToken := auth()

	res, err := getPlaylistItemInfo(singleId, accessToken)
	if err != nil {
		fmt.Println(err)
		return buildResponse(400, "Failed to load playlist info. Please check if playlist url valid. : ", nil), err
	}
	fmt.Println(res)

	return buildResponse(200, "Done", res), nil

}

func main() {
	lambda.Start(handler)
}
