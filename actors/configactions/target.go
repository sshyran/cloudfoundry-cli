package configactions

func (actor Actor) SetTarget(CCAPI string, skipSSLValidation bool) (string, Warnings, error) {
	warnings, err := actor.CloudControllerClient.TargetCF(CCAPI, skipSSLValidation)
	if err != nil {
		return "", Warnings(warnings), err
	}

	actor.Config.SetTargetInformation(
		actor.CloudControllerClient.API(),
		actor.CloudControllerClient.APIVersion(),
		actor.CloudControllerClient.AuthorizationEndpoint(),
		actor.CloudControllerClient.LoggregatorEndpoint(),
		actor.CloudControllerClient.DopplerEndpoint(),
		actor.CloudControllerClient.TokenEndpoint(),
	)

	return actor.CloudControllerClient.APIVersion(), Warnings(warnings), nil
}
