package base_test

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
	lokiURL    string
)

func TestBase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	lokiURL = integration.EnvOrDefault("LOKI_URL", "http://localhost:3100")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(webhookURL + "/eventreceiver/healthcheck")
	if err != nil || resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf(
			"Stack not running — webhook healthcheck at %s failed.\n"+
				"Start it with: task integration:up PROFILE=base\n",
			webhookURL+"/eventreceiver/healthcheck",
		)
		if err != nil {
			msg += fmt.Sprintf("Error: %v\n", err)
		}
		Fail(msg)
	}
	resp.Body.Close()
})
