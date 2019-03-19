[![Release](https://img.shields.io/github/release/Aptomi/k8s-app-engine.svg)](https://github.com/Aptomi/k8s-app-engine/releases/latest)
[![License](https://img.shields.io/github/license/Aptomi/k8s-app-engine.svg)](https://github.com/Aptomi/k8s-app-engine/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aptomi/k8s-app-engine)](https://goreportcard.com/report/github.com/Aptomi/k8s-app-engine)
[![Build Status](https://ci.aptomi.io/buildStatus/icon?job=aptomi%20-%20tests)](https://ci.aptomi.io/job/aptomi%20-%20tests/)
[![Coverage Status](https://coveralls.io/repos/github/Aptomi/k8s-app-engine/badge.svg)](https://coveralls.io/github/Aptomi/k8s-app-engine)
[![Godoc](https://godoc.org/github.com/Aptomi/k8s-app-engine?status.svg)](https://godoc.org/github.com/Aptomi/k8s-app-engine)
[![GitHub last commit](https://img.shields.io/github/last-commit/Aptomi/k8s-app-engine.svg)](https://github.com/Aptomi/k8s-app-engine/commits/master)
[![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io)

Engine for development teams that simplifies the roll-out and operation of container-based applications on **Kubernetes**. It introduces a service-centric abstraction, which allows users to compose applications from multiple connected components. It supports components packaged using Helm, ksonnet, k8s YAMLs, and any other Kubernetes-friendly way!

This approach to **application delivery** becomes especially powerful in a **multi-team** setup, where components owned by different teams must be put together into a service. With ownership boundaries, Dev teams can specify **multi-cluster** and **multi-env** (e.g. dev, stage, prod) service behavior, as well as control the lifecycle and updates of their respective services.

It also provides contextual **visibility** into teams and services, allowing our users to visualize complex dependencies and accurately assess the impact of changes. 

## Overview

[![What is](images/aptomi-overview-ng.png)](https://www.youtube.com/watch?v=j4DVtZvMqGg)

You can also read this [blog post](https://superuser.openstack.org/articles/aptomi-application-delivery-engine-k8s/) about the project.

## Demos

### Asciinema
* Installation - https://asciinema.org/a/163250
* Point to Kubernetes cluster - https://asciinema.org/a/163252
* Defining & running applications - https://asciinema.org/a/163258
* Updating a service instance - https://asciinema.org/a/163262
* Deleting service instances - https://asciinema.org/a/163269

### Youtube
* Detailed Demo - http://www.youtube.com/watch?v=HL4RwoBnuTc
* OpenDev CI/CD 2018 (Vancouver) - https://www.youtube.com/watch?v=GVB3kKocKi4 

## Table of contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Why?](#why)
- [Quickstart](#quickstart)
  - [Step #1: Installation](#step-1-installation)
  - [Step #2: Setting up k8s Cluster](#step-2-setting-up-k8s-cluster)
  - [Step #3: Running Examples](#step-3-running-examples)
  - [Step #4: Send us a note on Slack](#step-4-send-us-a-note-on-slack)
- [How It Works](#how-it-works)
  - [Architecture](#architecture)
  - [Language](#language)
- [How to contribute](#how-to-contribute)
- [Roadmap](#roadmap)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Why?

* Why would I want to use it [as a Developer?](docs/benefits.md#for-developers)
* Why would I want to use it [as an Operator?](docs/benefits.md#for-operators)
* How does it fit into [CI/CD model with Spinnaker/Jenkins?](docs/benefits.md#how-aptomi-fits-into-cicd-jenkins-spinnaker)
* Why would I want to use it [if I'm already implementing Kubernetes/OpenShift?](docs/benefits.md#why-would-i-want-to-use-aptomi-if-im-already-implementing-kubernetesopenshift)

## Quickstart

### Step #1: Installation
The simplest installation mode is **Compact**, but you may pick one that suits your needs:

Installation Mode     | Complexity | Description
----------------------|------------|-------------
[Compact](docs/install_compact.md) | Easy | Installation on local machine (binaries or in a single Docker container)
[Kubernetes](docs/install_kubernetes.md) | Medium | Deployed to k8s via Helm chart

You can also install it in a stripped-down mode with a fake executor:

Installation Mode     | Complexity | Description
----------------------|------------|-------------
[Concepts](docs/install_concepts.md) | Easy | Use this only if you want get familiar with the key concepts, API and UI. App deployment to k8s is **DISABLED**

### Step #2: Setting up k8s Cluster

You will need to have access to a k8s cluster to deploy apps from the provided examples.

Having a powerful k8s cluster with a good internet connection will definitely provide a *better experience* compared to a local, single-node k8s cluster. We consider GKE to be the best option if you don't have your own k8s cluster.

Kubernetes Cluster | When to use     | How to run
------------|-----------------|-----------
Your own    | If you already have a k8s cluster set up | [Configure on existing k8s cluster](docs/k8s_own.md)
Google Kubernetes Engine | If you have a Google account and free credits | [Configure on GKE](docs/k8s_gke.md)
k8s / Minikube | Single-node, local machine with 16GB+ RAM | [Configure on Minikube](docs/k8s_minikube.md)
k8s / Docker For Mac | Single-node, local machine with 16GB+ RAM | [Configure on Docker For Mac](docs/k8s_docker_for_mac.md)

### Step #3: Running Examples

Once the server is up and your k8s cluster is ready, you can get started by running the following examples:

Example    | Description
-----------|------------
[twitter-analytics](examples/twitter-analytics) | Twitter Analytics Application, multiple services, multi-cloud, based on Helm
[twitter-analytics-with-concourse-ci](examples/twitter-analytics-with-concourse-ci) | Twitter Analytics Application, integrated with Concourse CI/CD pipelines
[guestbook](examples/guestbook) | K8S Guestbook Application, multi-cloud, based on K8S YAMLs

### Step #4: Send us a note on Slack

Give us your feedback on #general in [![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io). If you run into any issues, we are always happy to help you resolve them! 

## How It Works

### Architecture
![Components](images/aptomi-components.png) 

See [artchitecture documentation](docs/architecture.md)

### Language
![Language](images/aptomi-language.png)

See [language documentation](docs/language.md)

## How to contribute
The very least you can do is [report a bug](https://github.com/Aptomi/k8s-app-engine/issues)!

If you want to make a pull request for a bug fix or contribute a feature, see our [Development Guide](docs/dev_guide.md) for how to develop, run and test your code.

We are always looking for feedback on how to make the project better. Join our Slack to discuss [![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io).

## Roadmap
[Feature Backlog](https://github.com/Aptomi/k8s-app-engine/milestone/11), as well as weekly project milestones, are good places to look at the roadmap items.

If you have any questions, please contact us on [![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io).
