package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"handler": "HelloHandler",
		"method":  r.Method,
		"path":    r.URL.Path,
		"remote":  r.RemoteAddr,
	}).Info("Received request")

	response := map[string]string{"message": "Hello from the Vibra backend!"}
	
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully sent response from HelloHandler")
}

func PresentHandler(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"handler": "PresentHandler",
		"method":  r.Method,
		"path":    r.URL.Path,
		"remote":  r.RemoteAddr,
	}).Info("Received request")

	livestreamData, repliesToLivestream, err := getPresentLivestream()
	if err != nil {
		logger.WithError(err).Error("Failed to get present livestream")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"livestreamData":      livestreamData,
		"repliesToLivestream": repliesToLivestream,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully sent response from PresentHandler")
}

func getPresentLivestream() (map[string]interface{}, []map[string]interface{}, error) {
	castHash := "0xbd78ba95ff14557be0a50746432df3cac0788758"
	url := fmt.Sprintf("https://api.neynar.com/v2/farcaster/cast/conversation?identifier=%s&type=hash&reply_depth=2&include_chronological_parent_casts=false&viewer_fid=16098&limit=50", castHash)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", os.Getenv("NEYNAR_API_KEY"))

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response body: %w", err)
	}

	var neynarResponse map[string]interface{}
	if err := json.Unmarshal(body, &neynarResponse); err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	conversation, ok := neynarResponse["conversation"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("conversation not found in response")
	}

	cast, ok := conversation["cast"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("cast not found in conversation")
	}

	author, ok := cast["author"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("author not found in cast")
	}

	reactions, ok := cast["reactions"].(map[string]interface{})
	if !ok {
		reactions = make(map[string]interface{})
	}

	replies, ok := cast["replies"].(map[string]interface{})
	if !ok {
		replies = make(map[string]interface{})
	}

	embeds, ok := cast["embeds"].([]interface{})
	if !ok || len(embeds) == 0 {
		return nil, nil, fmt.Errorf("no embeds found in cast")
	}

	embed := embeds[0].(map[string]interface{})
	streamUrl, ok := embed["url"].(string)
	if !ok {
		streamUrl = "Stream URL not found"
	}

	structuredResponse := map[string]interface{}{
		"castHash": cast["hash"],
		"streamer": map[string]interface{}{
			"fid":           author["fid"],
			"username":      author["username"],
			"display_name":  author["display_name"],
			"pfp_url":       author["pfp_url"],
			"followerCount": author["follower_count"],
		},
		"streamUrl":     streamUrl,
		"text":          cast["text"],
		"timestamp":     cast["timestamp"],
		"likes_count":   reactions["likes_count"],
		"recasts_count": reactions["recasts_count"],
		"replies_count": replies["count"],
		"channel": map[string]interface{}{
			"id":        cast["channel"].(map[string]interface{})["id"],
			"name":      cast["channel"].(map[string]interface{})["name"],
			"image_url": cast["channel"].(map[string]interface{})["image_url"],
		},
	}

	directReplies, ok := cast["direct_replies"].([]interface{})
	if !ok {
		return structuredResponse, []map[string]interface{}{}, nil
	}

	repliesToLivestream := make([]map[string]interface{}, 0, len(directReplies))
	for _, reply := range directReplies {
		replyMap, ok := reply.(map[string]interface{})
		if !ok {
			continue
		}
		
		replyAuthor, _ := replyMap["author"].(map[string]interface{})
		replyReactions, _ := replyMap["reactions"].(map[string]interface{})
		
		structuredReply := map[string]interface{}{
			"hash": replyMap["hash"],
			"author": map[string]interface{}{
				"fid":           replyAuthor["fid"],
				"username":      replyAuthor["username"],
				"display_name":  replyAuthor["display_name"],
				"pfp_url":       replyAuthor["pfp_url"],
			},
			"text":      replyMap["text"],
			"timestamp": replyMap["timestamp"],
			"reactions": map[string]interface{}{
				"likes_count":   replyReactions["likes_count"],
				"recasts_count": replyReactions["recasts_count"],
			},
		}
		repliesToLivestream = append(repliesToLivestream, structuredReply)
	}

	return structuredResponse, repliesToLivestream, nil
}


func extractStreamUrl(text string) string {
	// This is a placeholder function. You might want to implement
	// a more robust way to extract the stream URL from the cast text.
	return "https://www.youtube.com/watch?v=dZsIQV-B9Us"
}

func WriteCastHandler(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"handler": "WriteCastHandler",
		"method":  r.Method,
		"path":    r.URL.Path,
		"remote":  r.RemoteAddr,
	}).Info("Received request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		Text string `json:"text"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	castHash := "0xbd78ba95ff14557be0a50746432df3cac0788758"
	signerUUID := os.Getenv("ANKYSYNC_SIGNER_UUID")
	parentAuthorFid := 16098

	neynarResponse, err := writeCastToNeynar(requestBody.Text, signerUUID, parentAuthorFid, castHash)
	if err != nil {
		logger.WithError(err).Error("Failed to write cast to Neynar")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(neynarResponse)

	logger.Info("Successfully sent response from WriteCastHandler")
}

func writeCastToNeynar(text, signerUUID string, parentAuthorFid int, parentCastHash string) (map[string]interface{}, error) {
	url := "https://api.neynar.com/v2/farcaster/cast"

	payload := fmt.Sprintf(`{
		"signer_uuid": "%s",
		"text": "%s",
		"parent_author_fid": %d,
		"parent": "%s"
	}`, signerUUID, text, parentAuthorFid, parentCastHash)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", os.Getenv("NEYNAR_API_KEY"))
	req.Header.Add("content-type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var neynarResponse map[string]interface{}
	if err := json.Unmarshal(body, &neynarResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return neynarResponse, nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"handler": "LoginHandler",
		"method":  r.Method,
		"path":    r.URL.Path,
		"remote":  r.RemoteAddr,
	}).Info("Received login request")

	signedKey, err := neynar.GetSignedKey()
	if err != nil {
		logger.WithError(err).Error("Failed to get signed key")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signedKey)
}

func SignerStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"handler": "SignerStatusHandler",
		"method":  r.Method,
		"path":    r.URL.Path,
		"remote":  r.RemoteAddr,
	}).Info("Received signer status request")

	signerUUID := r.URL.Query().Get("signer_uuid")
	if signerUUID == "" {
		http.Error(w, "Missing signer_uuid parameter", http.StatusBadRequest)
		return
	}

	signer, err := neynar.LookupSigner(signerUUID)
	if err != nil {
		logger.WithError(err).Error("Failed to lookup signer")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signer)
}