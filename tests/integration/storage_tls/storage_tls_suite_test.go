package storage_tls_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var (
	webhookURL string
	s3Bucket   string
)

func TestStorageTLS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage TLS Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	s3Bucket = integration.EnvOrDefault("S3_BUCKET", "complybeacon-evidence")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(webhookURL + "/eventreceiver/healthcheck")
	if err != nil || resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf(
			"Stack not running — webhook healthcheck at %s failed.\n"+
				"Start it with: task integration:up PROFILE=storage-tls\n",
			webhookURL+"/eventreceiver/healthcheck",
		)
		if err != nil {
			msg += fmt.Sprintf("Error: %v\n", err)
		}
		Fail(msg)
	}
	resp.Body.Close()
})
