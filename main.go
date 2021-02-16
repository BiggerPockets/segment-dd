package main

import (
  "fmt"
  "os"
  "log"
  "bytes"
  "strings"
  "io/ioutil"
  "net/http"
  "crypto/hmac"
  "crypto/sha1"
  "encoding/hex"
  "regexp"
)

import (
  "gopkg.in/yaml.v2"
  "github.com/joho/godotenv"
  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
  "github.com/DataDog/datadog-go/statsd"
  "github.com/getsentry/sentry-go"
  sentrygin "github.com/getsentry/sentry-go/gin"
)

type Webhook struct {
  Type string `json:"type" binding:"required,oneof=track,identify"`
  Event string `json:"event"`
  UserID string `json:"userId"`
  Channel string `json:"channel"`
}

type Config struct {
  Events []string `yaml:""`
}

var statsdClient *statsd.Client
var statsdErr error
var sharedSecret string
var config Config

func main() {
  loadConfig()
  initializeStatsd()
  initializeSentry()
  router := initializeRouter()

  router.Run()
}

func initializeStatsd() {
  statsdClient, statsdErr = statsd.New("127.0.0.1:8125")

  if statsdErr != nil {
    fail(statsdErr)
    os.Exit(1)
  }
}

func loadConfig() {
  godotenv.Load()
  sharedSecret = os.Getenv("SEGMENT_SHARED_SECRET")
  error := yaml.Unmarshal(readConfig(), &config)

  if error != nil {
    fail(error)
    os.Exit(1)
  }
}

func readConfig() []byte {
  data, error := ioutil.ReadFile("config.yml")

  if error != nil {
    fail(error)
    os.Exit(1)
  }

  return data;
}

func initializeSentry() {
  error := sentry.Init(sentry.ClientOptions {
    Dsn: os.Getenv("SENTRY_DSN"),
    Environment: os.Getenv("GO_ENV"),
  })

  if error != nil {
    fail(error)
  }
}

func initializeRouter() *gin.Engine {
  router := gin.New();
  router.Use(gin.Logger());
  router.Use(gin.Recovery());

  router.Use(sentrygin.New(sentrygin.Options {
    Repanic: true,
  }));

  router.POST("/api/:source", processEvent)

  return router
}

func processEvent(context *gin.Context) {
  if validateRequest(context) == false {
    return
  }

  handleEvent(context, context.Param("source"))
}

func validateRequest(context *gin.Context) bool {
  signature := context.GetHeader("x-signature")

  if signature == "" {
    renderBadRequest(context)
    return false
  }

  requestBody := extractBody(context)

  if !validMAC(requestBody, signature) {
    renderUnauthorized(context)
    return false
  }

  return true
}

func extractBody(context *gin.Context) []byte {
  reader := context.Request.Body
  defer reader.Close()

  body, _ := ioutil.ReadAll(reader)
  context.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

  return body
}

func validMAC(message []byte, signature string) bool {
  messageMAC, _ := hex.DecodeString(signature)
  mac := hmac.New(sha1.New, []byte(sharedSecret))
  mac.Write([]byte(message))
  expectedMAC := mac.Sum(nil)

  return hmac.Equal(messageMAC, expectedMAC)
}

func renderBadRequest(context *gin.Context) {
  context.JSON(http.StatusBadRequest, gin.H {})
}

func renderUnauthorized(context *gin.Context) {
  context.JSON(http.StatusUnauthorized, gin.H {
    "error": "unauthorized",
  })
}

func renderError(context *gin.Context, err error) {
  context.JSON(http.StatusInternalServerError, gin.H {
    "error": err.Error(),
  })
}

func renderSuccess(context *gin.Context) {
  context.JSON(http.StatusOK, gin.H {})
}

func handleEvent(context *gin.Context, source string) {
  var webhook Webhook
  error := context.ShouldBindWith(&webhook, binding.JSON)

  if error != nil {
    renderError(context, error)
    return
  }

  if webhook.Type == "track" {
    handleTrackEvent(webhook, source)
    renderSuccess(context)
  }
}

func handleTrackEvent(webhook Webhook, source string) {
  if (!validEvent(webhook.Event)) {
    return
  }

  tags := []string {
    fmt.Sprintf("environment:%s", os.Getenv("GO_ENV")),
    fmt.Sprintf("source:%s", source),
    fmt.Sprintf("event:%s", webhook.Event),
    fmt.Sprintf("userId:%s", webhook.UserID),
    fmt.Sprintf("channel:%s", webhook.Channel),
    fmt.Sprintf("type:%s", webhook.Type),
  }

  eventName := formatEventName(webhook.Event)
  metricName := fmt.Sprintf("segment.event.%s", eventName)

  statsdClient.Incr(metricName, tags, 1)
}

func validEvent(eventName string) bool {
  for _, candidateEventName := range config.Events {
    if eventName == candidateEventName {
      return true
    }
  }

  return false
}

func formatEventName(eventName string) string {
  lowerEventName := strings.ToLower(eventName)
  regex := regexp.MustCompile("(?i)[^a-z0-9]")

  return regex.ReplaceAllString(lowerEventName, "_");
}

func fail(errorMessage error) {
  log.Fatalf("error: %v", errorMessage);
  sentry.CaptureException(errorMessage);
}
