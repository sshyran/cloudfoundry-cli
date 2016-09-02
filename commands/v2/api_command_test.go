package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/ui"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("API Command", func() {
	Context("when a valid api endpoint is specified", func() {
		var (
			cmd       ApiCommand
			UI        ui.UI
			fakeActor *v2fakes.FakeAPIConfigActor
		)

		BeforeEach(func() {
			out := NewBuffer()
			UI = ui.NewTestUI(out, out)
			fakeActor = new(v2fakes.FakeAPIConfigActor)

			cmd = ApiCommand{
				UI:    UI,
				Actor: fakeActor,
			}
		})

		Context("when passed an API URL with no protocol", func() {
			var err error
			JustBeforeEach(func() {
				err = cmd.Execute([]string{})
			})

			Context("when the API has SSL", func() {
				var CCAPI string
				BeforeEach(func() {
					CCAPI = "api.foo.com"
					cmd.OptionalArgs.URL = CCAPI

					fakeActor.SetTargetReturns("2.59.0", nil, nil)
				})

				Context("when the url has verified SSL", func() {
					It("sets the target", func() {
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
						url, skipSSLValidation := fakeActor.SetTargetArgsForCall(0)
						Expect(url).To(Equal("https://" + CCAPI))
						Expect(skipSSLValidation).To(BeFalse())

						Expect(UI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
						Expect(UI.Out).To(Say("OK"))
						Expect(UI.Out).To(Say("API endpoint:\\s+https://%s \\(API version: \\d+\\.\\d+\\.\\d+\\)", CCAPI))
					})
				})

				Context("when the url has unverified SSL", func() {
					Context("when --skip-ssl-validation is passed", func() {
						BeforeEach(func() {
							cmd.SkipSSLValidation = true
						})

						It("sets the target", func() {
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
							url, skipSSLValidation := fakeActor.SetTargetArgsForCall(0)
							Expect(url).To(Equal("https://" + CCAPI))
							Expect(skipSSLValidation).To(BeTrue())

							Expect(UI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
							Expect(UI.Out).To(Say("OK"))
							Expect(UI.Out).To(Say("API endpoint:\\s+https://%s \\(API version: \\d+\\.\\d+\\.\\d+\\)", CCAPI))
						})
					})

					Context("when no additional flags are passed", func() {
						BeforeEach(func() {
							fakeActor.SetTargetReturns("", nil, cloudcontrollerv2.UnverifiedServerError{})
						})

						It("returns an error with a --skip-ssl-validation tip", func() {
							Expect(err).To(MatchError(commands.FailedError{}))

							Expect(UI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
							Expect(UI.Out).To(Say("Invalid SSL Cert for %s", CCAPI))
							Expect(UI.Out).To(Say("TIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"))
						})
					})
				})
			})
		})

		Context("when passed an HTTP URL", func() {
			var CCAPI string

			BeforeEach(func() {
				CCAPI = "http://api.foo.com"
				cmd.OptionalArgs.URL = CCAPI

				fakeActor.SetTargetReturns("2.59.0", nil, nil)
			})

			It("sets the target with a warning", func() {
				err := cmd.Execute([]string{})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
				url, skipSSLValidation := fakeActor.SetTargetArgsForCall(0)
				Expect(url).To(Equal(CCAPI))
				Expect(skipSSLValidation).To(BeFalse())

				Expect(UI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
				Expect(UI.Out).To(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
				Expect(UI.Out).To(Say("OK"))
				Expect(UI.Out).To(Say("API endpoint:\\s+%s \\(API version: \\d+\\.\\d+\\.\\d+\\)", CCAPI))
			})
		})

		Context("when URL host does not exist", func() {
			var CCAPI string
			var expectedError error

			BeforeEach(func() {
				CCAPI = "i.do.not.exist.com"
				cmd.OptionalArgs.URL = CCAPI

				expectedError = cloudcontrollerv2.RequestError{Err: errors.New("I am an error")}
				fakeActor.SetTargetReturns("", nil, expectedError)
			})

			It("sets the target with a warning", func() {
				err := cmd.Execute([]string{})
				Expect(err).To(MatchError(commands.FailedError{}))

				Expect(UI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
				Expect(UI.Out).To(Say(expectedError.Error()))
				Expect(UI.Out).To(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
			})
		})
	})
})
