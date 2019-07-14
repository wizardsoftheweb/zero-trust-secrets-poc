# 09 The First Full Demo

I wanted to pause and review the goals I laid out in [the root `README`](./../../README.md#poc-onboarding). I've duplicated them here, because if I've met them, it's time to do more cool stuff.

## The Goals

I wanted to build something that...

1) provides an automated and centralized secret generation service,
2) provides different secrets to each environment while still remaining in the same pipeline,
3) provides a mechanism to update secrets in all environments automatically, and
4) provides a centralized location to monitor the status (not the contents) of the secrets.

###  Automated and Centralized Secrets

This is currently met in two parts. The control server is the primary hub of this PoC. With some development time, it would be very simple to add these things:

1) server authentication: only accept requests from trusted sources
2) config signatures: this should be easy to do but will require another library or two
3) myriad secrets: I feel like it's pretty obvious that changing things up won't be hard
4) scheduled updates: `cron` those jobs

The control server is not automated, though. Through the course of this PoC I've delegated most of that automation to the clients needing secrets. It makes more sense to me that the control server serves as a trusted generation point but doesn't know what it's generating. So far automating client secret generation is very easy.

### Different Secrets in Different Places

This is handily done. The secrets are so different there's no way to reproduce or replicate any of them. Every new thing that stands up can get its own unique secrets not seen or touched by anything else. As a huge proponent of good crypto and privacy on the web, I am stoked. As someone who might have to debug this later, I'm willing to admit there are a few things that should be touched up.

1) Audits: The initial keys should be passed off to another trusted, central location for reviewing config when things go wrong.
2) No replication: Secrets are unique to a single instance of any client. That means you can't replicate instances to increase throughput. If you're totally stateless, that's not a huge deal but there's a lot of other tooling that needs to be built to handle that. If you aren't stateless, you need replication at some point.
3) `gpg` is a huge pain in the ass: It was fast, simple, and got the MVP out the door. Dunno that I like it longterm.

### Automatic Updates

The resolution here is very similar to the first goal's resolution. The control server can reach out and update anything it's asked to with whatever values. A lot of that logic has been delegated out to the clients, at least for now. Using Viper, they can easily periodically update themselves. They'll also know if the remote config changes and they should adjust themselves automatically.

There is something to be said for some sort of secret orchestration solution on top of the control and clients. I think a necessary first step is figuring out how to share files such as the keys in such a way that the storage doesn't backdoor the encrypted config. It also requires some thought about what gets stored and how. If everything is totally ephemeral and unique, tasks such as scheduling key rotation become impossible because there's no way to track and organize them.

### Central Monitoring

This is the handwaviest resolution. I've got Prometheus metrics on everything. For the PoC, I didn't do much logging. I certainly didn't set things up well to automated recovery. To me, those things are part of actually committing to a project instead of trying to slap together something that works. I'm not deploying this anywhere yet so I've brushed those things off.

To some extent, this is also stuck in the same limbo as the previous two. To monitor status you need to track everything. You also need to store things. Figuring out the storage question and defining a good way to track all the moving pieces will give this goal a better foundation.

## The Demo

### Prereqs

You'll want [`kubectx`/`kubens`](https://github.com/ahmetb/kubectx).

These are assumed to be run before anything else.
```shell-session
minikube start --profile pocDemo --cpus 4 --memory 4096 
kubectx pocDemo
alias k=kubectl
```

