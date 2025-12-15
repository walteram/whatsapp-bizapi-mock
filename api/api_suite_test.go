package api_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	w_api "github.com/ron96G/whatsapp-bizapi-mock/api"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	log "github.com/ron96G/go-common-utils/log"
)

var (
	staticAPIToken = "abcdefg"
	apiPrefix      = "/v1"
	baseUrl        = "http://localhost:8080" + apiPrefix
	contacts       = []*model.Contact{}
	generators, _  = model.NewGenerators(w_api.Config.UploadDir, contacts, w_api.Config.InboundMedia)
	w              = webhook.NewWebhook(w_api.Config.ApplicationSettings.Webhooks.Url, w_api.Config.Version, generators)
	api            = w_api.NewAPI(apiPrefix, staticAPIToken, uint(20), w_api.Config, w)
	client         = StartNewServer(api.Server)

	marsheler = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
		Indent:       "  ",
	}

	unmarsheler = jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
)

func init() {
	log.Configure("debug", "json", os.Stdout)
}

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

func PanicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func StartNewServer(s *fasthttp.Server) (client *http.Client) {
	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := s.Serve(ln); err != nil {
			panic(err)
		}
	}()
	client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
		Timeout: time.Second,
	}

	return
}

var _ = Describe("Messages API", func() {
	defer GinkgoRecover()

	buf := bytes.NewBuffer(nil)

	AfterSuite(func() {
		err := api.Server.Shutdown()
		PanicIfNotNil(err)
	})

	Describe("Authentication", func() {

		Context("Missing Authorization Header", func() {
			message := &model.Message{
				To:   "+5511999999999",
				Type: model.MessageType_text,
				Text: &model.TextMessage{
					Body: "Hello World",
				},
			}
			buf.Reset()
			marsheler.Marshal(buf, message)
			req, _ := http.NewRequest("POST", baseUrl+"/123456789/messages", buf)
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 401", func() {
				Expect(resp.StatusCode).To(Equal(401))
			})
		})

		Context("Invalid API Key", func() {
			message := &model.Message{
				To:   "+5511999999999",
				Type: model.MessageType_text,
				Text: &model.TextMessage{
					Body: "Hello World",
				},
			}
			buf.Reset()
			marsheler.Marshal(buf, message)
			req, _ := http.NewRequest("POST", baseUrl+"/123456789/messages", buf)
			req.Header.Set("Authorization", "Bearer invalid_token")
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 401", func() {
				Expect(resp.StatusCode).To(Equal(401))
			})
		})

		Context("Valid API Key", func() {
			message := &model.Message{
				To:   "+5511999999999",
				Type: model.MessageType_text,
				Text: &model.TextMessage{
					Body: "Hello World",
				},
			}
			buf.Reset()
			marsheler.Marshal(buf, message)
			req, _ := http.NewRequest("POST", baseUrl+"/123456789/messages", buf)
			req.Header.Set("Authorization", "Bearer "+staticAPIToken)
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})

	Describe("Send Messages", func() {

		Context("Missing phoneNumberId", func() {
			message := &model.Message{
				To:   "+5511999999999",
				Type: model.MessageType_text,
				Text: &model.TextMessage{
					Body: "Hello World",
				},
			}
			buf.Reset()
			marsheler.Marshal(buf, message)
			req, _ := http.NewRequest("POST", baseUrl+"/messages", buf)
			req.Header.Set("Authorization", "Bearer "+staticAPIToken)
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 404", func() {
				Expect(resp.StatusCode).To(Equal(404))
			})
		})

		Context("Send Text Message", func() {
			message := &model.Message{
				To:   "+5511999999999",
				Type: model.MessageType_text,
				Text: &model.TextMessage{
					Body:       "Hello World",
					PreviewUrl: true,
				},
				MessagingProduct: "whatsapp",
			}
			buf.Reset()
			marsheler.Marshal(buf, message)
			req, _ := http.NewRequest("POST", baseUrl+"/123456789/messages", buf)
			req.Header.Set("Authorization", "Bearer "+staticAPIToken)
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})

			It("Should return a message ID", func() {
				idResp := new(model.IdResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, idResp))

				Expect(idResp.Messages).To(HaveLen(1))
				Expect(idResp.Messages[0].Id).ToNot(BeEmpty())
			})
		})

		Context("Send Image Message", func() {
			message := &model.Message{
				To:   "+5511999999999",
				Type: model.MessageType_image,
				Image: &model.ImageMessage{
					Id:      "1234567890",
					Link:    "https://example.com/image.jpg",
					Caption: "Check this out!",
				},
				MessagingProduct: "whatsapp",
			}
			buf.Reset()
			marsheler.Marshal(buf, message)
			req, _ := http.NewRequest("POST", baseUrl+"/123456789/messages", buf)
			req.Header.Set("Authorization", "Bearer "+staticAPIToken)
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})

}) // Messages API
