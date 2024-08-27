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
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type RequestBody struct {
	Type string   `json:"type"`
	Size int      `json:"size"`
	Urls []string `json:"urls"`
}

type ResponseBody struct {
	Status  int
	Message string
	Urls    []string
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

func extractImagesPlaylist(data map[string]interface{}) ([]string, error) {
	var images []string

	// "items" 추출
	items, ok := data["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("items field not found or incorrect type")
	}

	// 각 "item"을 순회
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// "track" 추출
		track, ok := itemMap["track"].(map[string]interface{})
		if !ok {
			continue
		}

		// "album" 추출
		album, ok := track["album"].(map[string]interface{})
		if !ok {
			continue
		}

		// "images" 추출
		imageList, ok := album["images"].([]interface{})
		if !ok {
			continue
		}

		// 각 이미지 정보를 Image 구조체로 변환하여 슬라이스에 추가
		for _, img := range imageList {
			imgMap, ok := img.(map[string]interface{})
			if !ok {
				continue
			}
			// 640 사이즈만 추가
			if int(imgMap["height"].(float64)) == 640 {
				images = append(images, imgMap["url"].(string))
			}

		}
	}

	return images, nil
}

func extractImagesTracks(data map[string]interface{}) ([]string, error) {
	var images []string

	// "items" 추출
	items, ok := data["tracks"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("tracks field not found or incorrect type")
	}

	// 각 "item"을 순회
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// "album" 추출
		album, ok := itemMap["album"].(map[string]interface{})
		if !ok {
			continue
		}

		// "images" 추출
		imageList, ok := album["images"].([]interface{})
		if !ok {
			continue
		}

		// 각 이미지 정보를 Image 구조체로 변환하여 슬라이스에 추가
		for _, img := range imageList {
			imgMap, ok := img.(map[string]interface{})
			if !ok {
				continue
			}
			// 640 사이즈만 추가
			if int(imgMap["height"].(float64)) == 640 {
				images = append(images, imgMap["url"].(string))
			}

		}
	}

	return images, nil
}

func getPlaylist(playlistID string, accessToken string) ([]string, error) {
	// API endpoint URL
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?fields=items(track(album(images)))", playlistID)
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

		images, err := extractImagesPlaylist(result)
		if err != nil {
			return nil, err
		}

		// Return the parsed JSON response
		return images, nil
	} else {
		// Print error message if request was not successful
		fmt.Printf("Error: %v\n", resp.Status)
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}
}

func getTracks(trackIDs []string, accessToken string) ([]string, error) {
	// API endpoint URL
	url := "https://api.spotify.com/v1/tracks?market=JP&ids=" + strings.Join(trackIDs, ",")
	fmt.Println(url)
	// Authorization header
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Send GET request to Spotify API
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %s", resp.Status)
	}

	// Parse JSON response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	images, err := extractImagesTracks(result)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func buildResponse(status int, msg string, urls []string) ResponseBody {
	return ResponseBody{
		Status:  status,
		Message: msg,
		Urls:    urls,
	}
}

func extractSpotifyObjectID(url, makingMethod string) string {
	// Check if "spotify" is in the URL
	if !strings.Contains(url, "spotify") {
		return url
	}

	// Remove query parameters by splitting at "?si"
	url = strings.Split(url, "?si")[0]

	// Define the regular expression pattern
	pattern := fmt.Sprintf(`https://open\.spotify\.com/%s/([a-zA-Z0-9]+)`, makingMethod)

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

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (ResponseBody, error) {

	var body RequestBody
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		log.Fatal(err)
	}

	ids := []string{}
	for _, url := range body.Urls {
		singleId := extractSpotifyObjectID(url, body.Type)
		if singleId == url {
			return buildResponse(400, "Invalid Url included : "+url, []string{}), nil
		}
		ids = append(ids, singleId)
	}

	accessToken := auth()

	if body.Type == "playlist" {
		res, err := getPlaylist(ids[0], accessToken)
		if err != nil {
			fmt.Println(err)
			return buildResponse(400, "Failed to load playlist info. Please check if playlist url valid. : "+ids[0], []string{}), err
		}
		fmt.Println(res)
		if len(res) < body.Size*body.Size {
			return buildResponse(400, "Playlist is too short", []string{}), nil
		}
		result := make([]string, body.Size*body.Size)
		perm := rand.Perm(len(res))
		for i, v := range perm[:body.Size*body.Size] {
			result[i] = res[v]
		}
		return buildResponse(200, "Done", result), nil
	}

	if body.Type == "track" {
		res, err := getTracks(ids, accessToken)
		if err != nil {
			fmt.Println(err)
			return buildResponse(400, "Failed to load track info. Please check if track url valid.", []string{}), err
		}
		fmt.Println(res)
		if len(res) < body.Size*body.Size {
			return buildResponse(400, "Not enough tracks", []string{}), nil
		}
		return buildResponse(200, "Done", res), nil
	}
	return buildResponse(500, "InternalError", []string{}), nil
}

func main() {
	lambda.Start(handler)
}
