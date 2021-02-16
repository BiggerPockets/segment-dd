package main

import "testing"
// import "net/http"
// import "net/http/httptest"
// import "net/url"

import "github.com/stretchr/testify/assert"

func TestFormatEventName(t *testing.T) {
	assert.Equal(t, "viewed_dashboard", formatEventName("Viewed Dashboard"))
}

func TestValidEvent(t *testing.T) {
	loadConfig();

	assert.True(t, validEvent("Viewed Dashboard"))
	assert.False(t, validEvent("Consumed Ice Cream"))
}

// func buildPayload() {
// 	data := url.Values{}
//   data.Set("name", "foo")
//   data.Set("surname", "bar")
// }

// func TestTrackEvent(t *testing.T) {
// 	router := initializeRouter()
// 	recorder := httptest.NewRecorder()

// 	request, _ := http.NewRequest("POST", "/api/test", nil)
// 	request.Header.Add("x-signature", "INVALIDSIGNATURE")
// 	router.ServeHTTP(recorder, request)

// 	assert.Equal(t, 200, recorder.Code)
// 	assert.Equal(t, "pong", recorder.Body.String())
// }

// func TestTrackEventWithoutSignature(t *testing.T) {
// 	router := initializeRouter()
// 	recorder := httptest.NewRecorder()

// 	request, _ := http.NewRequest("POST", "/api/test", nil)
// 	router.ServeHTTP(recorder, request)

// 	assert.Equal(t, 400, recorder.Code)
// }

// func TestTrackEventWitInvalidSignature(t *testing.T) {
// 	router := initializeRouter()
// 	recorder := httptest.NewRecorder()

// 	request, _ := http.NewRequest("POST", "/api/test", nil)
// 	request.Header.Add("x-signature", "INVALIDSIGNATURE")
// 	router.ServeHTTP(recorder, request)

// 	assert.Equal(t, 401, recorder.Code)
// }
