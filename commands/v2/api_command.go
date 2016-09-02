package v2

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actors/configactions"
	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

//go:generate counterfeiter . APIConfigActor

type APIConfigActor interface {
	SetTarget(CCAPI string, skipSSLValidation bool) (string, configactions.Warnings, error)
}

type ApiCommand struct {
	OptionalArgs      flags.APITarget `positional-args:"yes"`
	SkipSSLValidation bool            `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	Unset             bool            `long:"unset" description:"Remove all api endpoint targeting"`
	usage             interface{}     `usage:"CF_NAME api [URL]"`
	relatedCommands   interface{}     `related_commands:"auth, login, target"`

	UI    UI
	Actor APIConfigActor
}

func (cmd ApiCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.Actor = configactions.NewActor(config, cloudcontrollerv2.NewCloudControllerClient())
	return nil
}

func (cmd ApiCommand) Execute(args []string) error {
	cmd.UI.DisplayHeaderFlavorText("Setting api endpoint to {{.API}}...", map[string]interface{}{
		"API": cmd.OptionalArgs.URL,
	})

	api := cmd.processURL(cmd.OptionalArgs.URL)

	apiVersion, _, err := cmd.Actor.SetTarget(api, cmd.SkipSSLValidation)
	if err != nil {
		return cmd.handleError(err)
	}

	if strings.HasPrefix(api, "http:") {
		cmd.UI.DisplayText("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended")
	}

	cmd.UI.DisplayText("OK")
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("API endpoint: {{.APIEndpoint}} (API version: {{.APIVersion}})", map[string]interface{}{
		"APIEndpoint": api,
		"APIVersion":  apiVersion,
	})

	return nil
}

func (_ ApiCommand) processURL(apiURL string) string {
	if !strings.HasPrefix(apiURL, "http") {
		return fmt.Sprintf("https://%s", apiURL)

	}
	return apiURL
}

func (cmd ApiCommand) handleError(err error) error {
	switch e := err.(type) {
	case cloudcontrollerv2.UnverifiedServerError:
		cmd.UI.DisplayErrorMessage(
			"Invalid SSL Cert for {{.API}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint",
			map[string]interface{}{
				"API": cmd.OptionalArgs.URL,
			})

	case cloudcontrollerv2.RequestError:
		cmd.UI.DisplayErrorMessage(e.Error())
		cmd.UI.DisplayErrorMessage("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection.")

	default:
		cmd.UI.DisplayText(err.Error())
	}
	return commands.FailedError{}

}
