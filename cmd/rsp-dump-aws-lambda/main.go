package main

import (
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"github.com/CursedHardware/go-rsp-dump/rsp/dump"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"gopkg.in/mail.v2"
	"log"
	"net/http"
	"os"
	"regexp"
)

var config = Configuration{
	Homepage:    "https://septs.blog/posts/rsp-dump/",
	HostPattern: regexp.MustCompile(`^(?P<issuer>[a-f0-9]{6,40})\.rsp\.`),
	SMTPPort:    587,
	SMTPHeaders: make(map[string][]string),
}

var smtpClient *mail.Dialer

func init() {
	if fp, err := os.Open("rsp-config.json"); err != nil {
		log.Fatalln(err)
	} else if err = json.NewDecoder(fp).Decode(&config); err != nil {
		log.Fatalln(err)
	}
	if _, ok := config.SMTPHeaders["From"]; !ok {
		config.SMTPHeaders["From"] = []string{config.SMTPUsername}
	}
	smtpClient = mail.NewDialer(config.SMTPHost, int(config.SMTPPort), config.SMTPUsername, config.SMTPPassword)
	smtpClient.StartTLSPolicy = mail.MandatoryStartTLS
}

func main() {
	log.SetFlags(0)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	handler := &dump.Handler{
		Homepage:       config.Homepage,
		Client:         http.DefaultClient,
		Issuers:        mustRSPRegistry(),
		HostPattern:    config.HostPattern,
		OnAuthenClient: onAuthenClient,
	}
	lambda.Start(httpadapter.New(handler).ProxyWithContext)
}

func mustRSPRegistry() (issuers map[string][]string) {
	fp, err := os.Open("rsp-registry.json")
	if err != nil {
		panic(err)
	}
	if err = json.NewDecoder(fp).Decode(&issuers); err != nil {
		panic(err)
	}
	return
}
