# Barbossa

[![Build Status](https://travis-ci.org/jelmersnoeck/barbossa.svg?branch=master)](https://travis-ci.org/jelmersnoeck/barbossa)

Barbossa's aim is to help out take care of the safety and security of your
Kubernetes Resources, much like a [Chief Mate](https://en.wikipedia.org/wiki/Chief_mate).

It does so by setting up several policies and presets to make sure that the
applications you deploy in your Kubernetes Cluster are configured to be Highly
Available.

## Installation

To install the webhook, you'll need to set up a `CertificateSigningRequest`,
this can be done by running the `hack/webhook-create-signed-cert.sh` script.

```
kubectl apply -f docs/kube/00-namespace.yaml
./hack/webhook-create-signed-cert.sh
```

This will configure a new certificate which will allow communication between the
APIServer and the Webhook.

Next, you'll want to inject the generated certificate in the Webhook
configuration before it's applied to your cluster. To do so, you can run the
following:

```
cat docs/kube/with-rbac.yaml | ./hack/webhook-patch-ca-bundle.sh | kubectl apply -f -
```

This will read the generated certificates from the configuration file, inject it
into the Webhook configuration and install it on your cluster.

**Note:** we'll look at allowing certificates to be issued by cert-manager and
using the APIService methodology to configure this instead of communicating
straight with the Service.
