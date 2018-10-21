# Barbossa

[![Build Status](https://travis-ci.org/jelmersnoeck/barbossa.svg?branch=master)](https://travis-ci.org/jelmersnoeck/barbossa)

Barbossa's aim is to help out take care of the safety and security of your
Kubernetes Resources, much like a [Chief Mate](https://en.wikipedia.org/wiki/Chief_mate).

It does so by setting up several policies and presets to make sure that the
applications you deploy in your Kubernetes Cluster are configured to be Highly
Available.

## Installation

To install the webhook, you'll need [cert-manager](https://github.com/jetstack/cert-manager) installed in your cluster.

Once that's done, you can install the webhook as follows:

```
kubectl apply -f docs/kube
```

## Defaults

To achieve High Availability within a Kubernetes cluster, we've configured some
defaults to enforce setting up some values. You can view these configurations in
the [defaults](./docs/kube/defaults) folder.
