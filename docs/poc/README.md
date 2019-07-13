# PoC Runthrough

Docs get added here as I write them.

## Current Stuff

1. [Configuring local `etcd`](./01-etcd.md): `etcd` needs to be running somewhere.
2. [Working with `crypt`](./02-crypt.md): `crypt` is a simple package that can encrypt config values in `etcd` using a GPG key/
3. [Creating a control server](./03-control-server.md): The control server creates secrets, encrypts them, and sends them to `etcd`.
4. [Building a basic client](./04-simple-client.md): The first iteration of the client creates a GPG keyring, adds some secrets to a remote config in `etcd`, and updates the secrets as the remote config changes. 
5. [Terraform provider tangent](./05-terraform-provider.md): A custom provider can be created to interact with the control server.
6. [GPG in a container](./06-client-container.md): Clients need `gpg`
7. [Starter Kube](./07-kubernetes.md): Everything is containerized and running in a kube cluster
