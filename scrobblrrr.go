package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/sideb0ard/mack"
)

const (
	APIKEY    = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	AUTHTOKEN = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	SECRIT    = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	SK        = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	WSAPI     = "http://ws.audioscrobbler.com/2.0/"
)

func hashy(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func signCall(paramstring string) string {
	re, _ := regexp.Compile(`=`)
	tidyparams := re.ReplaceAllString(paramstring, "")
	params := strings.Split(tidyparams, "&")
	// fmt.Println(params)
	var buffer bytes.Buffer
	sort.Strings(params)
	for _, v := range params {
		buffer.WriteString(v)
	}
	buffer.WriteString(SECRIT)
	//fmt.Println("signature:", buffer.String())

	hashy := hashy(buffer.String())
	return hashy
}

func getAuthToken() string {
	// siggy := signCall([]string{"methodauth.gettoken", "api_key63ca05fc03519ad79c76909e36b7926e"})
	params := fmt.Sprintf("?method=auth.gettoken&api_key=%s&format=json", APIKEY)
	urlstring := fmt.Sprintf("%s%s", WSAPI, params)
	resp, err := http.Get(urlstring)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// status := resp.Status
	var f interface{}
	err = json.Unmarshal(body, &f)

	m := f.(map[string]interface{})
	for k, v := range m {
		if k == "token" {
			return v.(string)
		}
		// fmt.Println(k, v)
	}
	//return string(fmt.Sprintf("%s - %s", status, body))
	return "blah"

}

func getTrackData() (artist string, album string, name string) {
	// track, _ = mack.Tell("iTunes", `set current_track to the current track`)
	artist, _ = mack.Tell("iTunes", `set track_artist to the artist of the current track`)
	album, _ = mack.Tell("iTunes", `set track_album to the album of the current track`)
	name, _ = mack.Tell("iTunes", `set track_name to the name of the current track`)
	return
}

func getUserAuth(authToken string) {
	urly := fmt.Sprintf("http://www.last.fm/api/auth/?api_key=%s&token=%s", APIKEY, authToken)
	browser.OpenURL(urly)
}

func setNowPlaying(artistName string, albumName string, trackName string) {
	method := "track.updateNowPlaying"
	params := fmt.Sprintf("method=%s&artist=%s&track=%s&api_key=%s&sk=%s", method, artistName, trackName, APIKEY, SK)
	sig := signCall(params)
	data := bytes.NewBufferString(params + "&api_sig=" + sig)
	// fmt.Println(data)
	// r, _ := http.NewRequest("POST", WSAPI, bytes.NewBufferString(data.Encode()))
	resp, err := http.Post(WSAPI, "application/x-www-form-urlencoded", data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(time.Now().String(), resp)
}

func scrobbleTrack(artistName string, albumName string, trackName string) {
	method := "track.scrobble"
	now := strconv.FormatInt(time.Now().Unix(), 10)
	params := fmt.Sprintf("method=%s&artist=%s&track=%s&api_key=%s&sk=%s&timestamp=%s", method, artistName, trackName, APIKEY, SK, now)
	sig := signCall(params)
	data := bytes.NewBufferString(params + "&api_sig=" + sig)
	// fmt.Println(data)
	// r, _ := http.NewRequest("POST", WSAPI, bytes.NewBufferString(data.Encode()))
	resp, err := http.Post(WSAPI, "application/x-www-form-urlencoded", data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(time.Now().String(), resp)
}

func getSessionKey() {
	method := "auth.getSession"
	params := fmt.Sprintf("method=%s&api_key=%s&token=%s", method, APIKEY, AUTHTOKEN)
	sig := signCall(params)
	fmt.Println(sig)
	urly := WSAPI + "?" + params + "&api_sig=" + sig
	fmt.Println(urly)
	// resp, err := http.Get("http://example.com/")
}

func main() {
	// authToken := getAuthToken()
	// fmt.Println("authytoken", authToken)
	//getUserAuth(authToken)
	// getSessionKey()
	// fmt.Println(sessionToken)
	previousArtist := ""
	previousTrack := ""
	for {
		fmt.Println(time.Now(), " - Aiiight, scrobble time..")
		artist, album, trackTitle := getTrackData()
		if artist == previousArtist && trackTitle == previousTrack {
			fmt.Println(time.Now(), " - Ah, still same track - back to sleep..")
			// nuthing
		} else {
			setNowPlaying(artist, album, trackTitle)
			fmt.Printf("%s Scrobble:: artist: %s, track: %s\n", time.Now().String(), artist, trackTitle)
			scrobbleTrack(artist, album, trackTitle)
			previousArtist = artist
			previousTrack = trackTitle
		}
		time.Sleep(30 * time.Second)
	}
}
