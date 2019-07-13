# 06 Client Container

If the control server requires a GPG pub key to encrypt config, the clients need to be able to manage GPG keys. For this stage, I'd like to build a minimal container using an Ubuntu LTS image that clients can be dropped into. The container needs the following:

* `gpg2` must be on the box
* The primary keyring must be built automatically
* The container should expose uid and gid for a new user to use the keyring so clients can be built with service users instead of running as `root`
