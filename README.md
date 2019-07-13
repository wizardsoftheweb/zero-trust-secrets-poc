# Zero Trust Secrets

The goal of this project is to combine a bunch of really neat stuff into a broke dev's security at rest solution.

## Very Important Caveat

Humans are the worst part of any security system. Bad decisions are by even the most intelligent people. No security is perfect.

At this stage in the project I'm completely ignoring normal authentication and flow security. I assume production services should have that already. In other words, this isn't yet a production service and you shouldn't treat it as one.

## PoC Onboarding

I'd going to attempt to tie these things together into something resembling a zero trust secret service. That is, no secrets are stored unencrypted, no secrets are passed unencrypted, secrets are generated automatically per environment with an audit trail of access, and, most importantly, the end user most likely never sees the secrets. Those are my goals, at least.

This list is rather small because I'm still in the planning phase. MVP will probably be a web server with a k8s or TF component and a CLI.

This first phase is pretty small. I want to get a quick MVP out that 

1) provides an automated and centralized secret generation service,
2) provides different secrets to each environment while still remaining in the same pipeline,
3) provides a mechanism to update secrets in all environments automatically, and
4) provides a centralized location to monitor the status (not the contents) of the secrets.

For a first pass, this is the tech I think I'm using.

* [crypt](https://github.com/xordataexchange/crypt): this tool uses OpenPGP to lock your data down, then sets it up to be shipped to a storage provider
* [etcd](https://github.com/etcd-io/etcd): tbh I've never used either of the providers that crypt supports OOTB so I picked the one that wasn't HashiCorp
* [Viper](https://github.com/spf13/viper): this is essentially a configuration broker; its integration with crypt is what set this off
* [Cobra](https://github.com/spf13/cobra): same dude that built Viper which means the CLI integration is baked in
* [Gin](https://github.com/gin-gonic/gin): seems easy and fast

I've also got a stretch goal if I find the time. This is something that could be useful for Terraform, so making a provider is on my radar.

* [Custom Providers](https://www.terraform.io/docs/extend/writing-custom-providers.html): creating secrets directly during the planning/application phase is ideal
* [`terraform-provider-vagrant`](https://github.com/bmatcuk/terraform-provider-vagrant): why the hell not
* [HTTP backend](https://www.terraform.io/docs/backends/types/http.html): TF can't and won't encrypt the backend for you; a loopback webservice that encrypts on its way out might do the trick

### Full PoC Run Through

It's being written [here](./docs/poc/README.md).
