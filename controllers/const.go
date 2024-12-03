package controllers

const (
	OPERATOR_SEED_KEY    = "seed.nk"
	OPERATOR_PUBLIC_KEY  = "key.pub"
	OPERATOR_JWT         = "key.jwt"
	OPERATOR_CREDS       = "user.creds"
	OPERATOR_CONFIG_FILE = "auth.conf"
	AUTH_CONFIG_TEMPLATE = `operator: %s
system_account: %s
resolver {
	type: full
	dir: './jwt'
	allow_delete: true
	interval: "2m"
	timeout: "5s"
}
resolver_preload: {
	%s: %s,
}
`
)
