package vars

var (
	// SupervisorNamespace is the supervisor service namespace.
	SupervisorNamespace = "pinniped-supervisor"

	// SupervisorSvcName is the supervisor service name.
	SupervisorSvcName = ""

	// ConciergeNamespace is the Concierge namespace.
	ConciergeNamespace = "pinniped-concierge"

	// DexNamespace is the Dex namespace.
	DexNamespace = "tanzu-system-auth"

	// DexSvcName is the Dex service name.
	DexSvcName = "dexsvc"

	// DexCertName is the Dex cert name.
	DexCertName = "dex-cert-tls"

	// DexConfigMapName is the Dex ConfigMap name.
	DexConfigMapName = "dex"

	// PinnipedOIDCProviderName is the Dex OIDC identity provider.
	PinnipedOIDCProviderName = "dex-oidc-identity-provider"

	// PinnipedOIDCClientSecretName is the Dex client credentials.
	PinnipedOIDCClientSecretName = "dex-client-credentials"

	// SupervisorSvcEndpoint is the supervisor service endpoint.
	SupervisorSvcEndpoint = ""

	// FederationDomainName is the federation domain name.
	FederationDomainName = ""

	// JWTAuthenticatorName is the JWT authentication name.
	JWTAuthenticatorName = ""

	// JWTAuthenticatorAudience is the JWT authentication audience.
	JWTAuthenticatorAudience = ""

	// SupervisorCertName is the supervisor cert name.
	SupervisorCertName = ""

	// SupervisorCABundleData is the supervisor CA bundle data.
	SupervisorCABundleData = ""
)
