package WxSend

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileType string

const (
	Image FileType = "image"
	Voice FileType = "voice"
	Video FileType = "video"
	File  FileType = "file"
)

type TypeStruct struct {
	Type          FileType
	MediaID       string
	VideoSettings struct {
		Title       string
		Description string
	}
}

// GetToken
//
// @arg: `corpid` (string)
//
// @arg: `corpsecret` (string)
//
// @return: `access_token` (string)
//
// @return: `expire_in` (int64)
//
// @return: `err` (error)
//
func GetToken(CorpID, CorpSecret string) (string, int64, error) {
	URL := &url.URL{
		Scheme: "https",
		Host:   "qyapi.weixin.qq.com",
		Path:   "/cgi-bin/gettoken",
		RawQuery: func() string {
			ParamsMap := make(map[string]string)
			ParamsMap["corpid"] = CorpID
			ParamsMap["corpsecret"] = CorpSecret
			ParamsSlice := make([]string, 0)
			for k, v := range ParamsMap {
				ParamsSlice = append(ParamsSlice, k+"="+v)
			}
			return strings.Join(ParamsSlice, "&")
		}(),
	}
	Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
		Timeout: 3 * time.Second,
	}
	Request, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return "", 0, errors.New("create request fail: " + err.Error())
	}
	Response, err := Client.Do(Request)
	if err != nil {
		return "", 0, errors.New("get response fail: " + err.Error())
	}
	if Response.StatusCode != 200 {
		return "", 0, errors.New("fail to get body, status code: " + Response.Status)
	}
	DataRaw, err := ioutil.ReadAll(Response.Body)
	if err != nil {
		return "", 0, errors.New("fail to get body: " + err.Error())
	}
	type ResponseBodyStruct struct {
		ErrCode     int64  `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiredTime int64  `json:"expires_in"`
	}
	var ResponseJson ResponseBodyStruct
	err = json.Unmarshal(DataRaw, &ResponseJson)
	if err != nil {
		return "", 0, errors.New("json parse fail: " + err.Error())
	}
	if ResponseJson.ErrCode != 0 {
		return "", 0, errors.New(ResponseJson.ErrMsg)
	}
	return ResponseJson.AccessToken, ResponseJson.ExpiredTime, nil
}

func SendText(AccessToken, AppID, Username string, Data []byte) error {
	URL := &url.URL{
		Scheme: "https",
		Host:   "qyapi.weixin.qq.com",
		Path:   "/cgi-bin/message/send",
		RawQuery: func() string {
			ParamsMap := make(map[string]string)
			ParamsMap["access_token"] = AccessToken
			ParamsSlice := make([]string, 0)
			for k, v := range ParamsMap {
				ParamsSlice = append(ParamsSlice, k+"="+v)
			}
			return strings.Join(ParamsSlice, "&")
		}(),
	}
	Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
		Timeout: 3 * time.Second,
	}
	type RequestBodyStruct struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		AppID   string `json:"agentid"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
		EnableDuplicateCheck   int `json:"enable_duplicate_check"`
		DuplicateCheckInterval int `json:"duplicate_check_interval"`
	}
	RequestBody := RequestBodyStruct{
		ToUser:  Username,
		MsgType: "text",
		AppID:   AppID,
		Text: struct {
			Content string `json:"content"`
		}{
			Content: string(Data),
		},
		EnableDuplicateCheck:   0,
		DuplicateCheckInterval: 1800,
	}
	RequestBodyBytes, err := json.Marshal(RequestBody)
	if err != nil {
		return errors.New("json parse fail: " + err.Error())
	}
	RequestBodyIOReader := bytes.NewReader(RequestBodyBytes)
	Request, err := http.NewRequest(http.MethodPost, URL.String(), RequestBodyIOReader)
	if err != nil {
		return errors.New("create request fail: " + err.Error())
	}
	Response, err := Client.Do(Request)
	if err != nil {
		return errors.New("get response fail: " + err.Error())
	}
	if Response.StatusCode != 200 {
		return errors.New("fail to get body, status code: " + Response.Status)
	}
	DataRaw, err := ioutil.ReadAll(Response.Body)
	if err != nil {
		return errors.New("fail to get body: " + err.Error())
	}
	type ResponseBodyStruct struct {
		ErrCode     int64  `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		Type        string `json:"type"`
		MediaID     string `json:"media_id"`
		CreatedTime string `json:"created_at"`
	}
	var ResponseJson ResponseBodyStruct
	err = json.Unmarshal(DataRaw, &ResponseJson)
	if err != nil {
		return errors.New("json parse fail: " + err.Error())
	}
	if ResponseJson.ErrCode != 0 {
		return errors.New(ResponseJson.ErrMsg)
	}
	return nil
}

func SendFile(AccessToken, AppID, Username string, TypeSettings TypeStruct) error {
	URL := &url.URL{
		Scheme: "https",
		Host:   "qyapi.weixin.qq.com",
		Path:   "/cgi-bin/message/send",
		RawQuery: func() string {
			ParamsMap := make(map[string]string)
			ParamsMap["access_token"] = AccessToken
			ParamsSlice := make([]string, 0)
			for k, v := range ParamsMap {
				ParamsSlice = append(ParamsSlice, k+"="+v)
			}
			return strings.Join(ParamsSlice, "&")
		}(),
	}
	Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
		Timeout: 3 * time.Second,
	}
	RequestBodyMap := make(map[string]interface{})
	RequestBodyMap["touser"] = Username
	RequestBodyMap["agentid"] = AppID
	switch TypeSettings.Type {
	case Image:
		RequestBodyMap["msgtype"] = "image"
		RequestBodyMap["image"] = map[string]string{
			"media_id": TypeSettings.MediaID,
		}
	case Voice:
		RequestBodyMap["msgtype"] = "voice"
		RequestBodyMap["voice"] = map[string]string{
			"media_id": TypeSettings.MediaID,
		}
	case Video:
		RequestBodyMap["msgtype"] = "video"
		if TypeSettings.VideoSettings.Title == "" {
			TypeSettings.VideoSettings.Title = "Video_" + strconv.FormatInt(time.Now().Unix(), 10)
		}
		RequestBodyMap["video"] = map[string]string{
			"media_id":    TypeSettings.MediaID,
			"title":       TypeSettings.VideoSettings.Title,
			"description": TypeSettings.VideoSettings.Description,
		}
	case File:
		RequestBodyMap["msgtype"] = "file"
		RequestBodyMap["file"] = map[string]string{
			"media_id": TypeSettings.MediaID,
		}
	default:
		return errors.New("invalid type")
	}
	RequestBodyMap["enable_duplicate_check"] = 0
	RequestBodyMap["duplicate_check_interval"] = 1800
	RequestBodyBytes, err := json.Marshal(RequestBodyMap)
	if err != nil {
		return errors.New("json parse fail: " + err.Error())
	}
	RequestIOReader := bytes.NewReader(RequestBodyBytes)
	Request, err := http.NewRequest(http.MethodPost, URL.String(), RequestIOReader)
	if err != nil {
		return errors.New("create request fail: " + err.Error())
	}
	Response, err := Client.Do(Request)
	if err != nil {
		return errors.New("get response fail: " + err.Error())
	}
	if Response.StatusCode != 200 {
		return errors.New("fail to get body, status code: " + Response.Status)
	}
	DataRaw, err := ioutil.ReadAll(Response.Body)
	if err != nil {
		return errors.New("fail to get body: " + err.Error())
	}
	type ResponseBodyStruct struct {
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
		InvalidUser  string `json:"invaliduser"`
		MsgId        string `json:"msgid"`
		ResponseCode string `json:"response_code"`
	}
	var ResponseBody ResponseBodyStruct
	err = json.Unmarshal(DataRaw, &ResponseBody)
	if err != nil {
		return errors.New("json parse fail: " + err.Error())
	}
	if ResponseBody.ErrCode != 0 {
		return errors.New(ResponseBody.ErrMsg)
	}
	return nil
}

func UploadFileFromStream(AccessToken string, Stream io.Reader, Filename string, Type FileType) (string, error) {
	_, FileNameSmall := filepath.Split(Filename)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="media"; filename="%s"`, FileNameSmall))
	h.Set("Content-Type", "application/octet-stream")
	fileWriter, err := bodyWriter.CreatePart(h)
	if err != nil {
		return "", errors.New("file read prepare fail: " + err.Error())
	}
	_, err = io.Copy(fileWriter, Stream)
	if err != nil {
		return "", errors.New("io copy fail: " + err.Error())
	}
	ContentType := bodyWriter.FormDataContentType()
	_ = bodyWriter.Close()
	Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
		Timeout: 3 * time.Second,
	}
	URL := &url.URL{
		Scheme: "https",
		Host:   "qyapi.weixin.qq.com",
		Path:   "/cgi-bin/media/upload",
		RawQuery: func() string {
			ParamsMap := make(map[string]string)
			ParamsMap["access_token"] = AccessToken
			ParamsMap["type"] = string(Type)
			ParamsSlice := make([]string, 0)
			for k, v := range ParamsMap {
				ParamsSlice = append(ParamsSlice, k+"="+v)
			}
			return strings.Join(ParamsSlice, "&")
		}(),
	}
	Request, err := http.NewRequest(http.MethodPost, URL.String(), bodyBuf)
	if err != nil {
		return "", errors.New("create request fail: " + err.Error())
	}
	Request.Header.Set("Content-Type", ContentType)
	Response, err := Client.Do(Request)
	if err != nil {
		return "", errors.New("get response fail: " + err.Error())
	}
	if Response.StatusCode != 200 {
		return "", errors.New("fail to get body, status code: " + Response.Status)
	}
	DataRaw, err := ioutil.ReadAll(Response.Body)
	if err != nil {
		return "", errors.New("fail to get body: " + err.Error())
	}
	type ResponseBodyStruct struct {
		ErrCode     int64  `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		Type        string `json:"type"`
		MediaID     string `json:"media_id"`
		CreatedTime string `json:"created_at"`
	}
	var ResponseJson ResponseBodyStruct
	err = json.Unmarshal(DataRaw, &ResponseJson)
	if err != nil {
		return "", errors.New("json parse fail: " + err.Error())
	}
	if ResponseJson.ErrCode != 0 {
		return "", errors.New(ResponseJson.ErrMsg)
	}
	return ResponseJson.MediaID, nil
}
