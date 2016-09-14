package v2_test

import (
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	"code.cloudfoundry.org/cli/commands/ui"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"
	"code.cloudfoundry.org/cli/utils/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = XDescribe("Unbind Service Command", func() {
	var (
		cmd        UnbindServiceCommand
		fakeUI     ui.UI
		fakeActor  *v2fakes.FakeUnbindServiceActor
		fakeConfig *commandsfakes.FakeConfig
	)

	BeforeEach(func() {
		out := NewBuffer()
		fakeUI = ui.NewTestUI(out, out)
		fakeActor = new(v2fakes.FakeUnbindServiceActor)
		fakeConfig = new(commandsfakes.FakeConfig)

		cmd = UnbindServiceCommand{
			UI:    fakeUI,
			Actor: fakeActor,
		}
		fakeConfig.TargetedOrganizationReturns(config.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(config.Space{
			Name: "some-space",
		})
		fakeConfig.CurrentUserReturns(config.User{
			Name: "admin",
		}, nil)
	})

	Context("when both the app and service exist", func() {
		var err error

		BeforeEach(func() {

		})

		JustBeforeEach(func() {
			err = cmd.Execute([]string{})
		})

		Context("when a binding exists between the app and the service", func() {
			It("successfully unbinds the app", func() {
			})
		})

		Context("when no binding exists between the app and the service", func() {
			It("states the binding does not exist", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("Unbinding app %s from service %s in org %s / space %s as %s",
					"some-app",
					"some-service",
					"some-org",
					"some-space",
					"some-user",
				))
				Expect(fakeUI.Out).To(Say("OK"))
				Expect(fakeUI.Out).To(Say("Binding between %s and %s did not exist", "some-app", "some-service"))
			})
		})
	})

	Context("when the service does not exist", func() {
	})

	Context("when the app does not exist", func() {
	})
})
