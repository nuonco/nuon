# Basic component

## At a glance

In this guide you’ll learn how to configure and deploy a basic component using either a public container image or a Dockerfile inside one of your GitHub repositories.

## Create a component

First you’ll need to go to the components page and click the “Add component” button. The component configuration panel will open and you can start configuring your component.

*Note: currently Nuon only supports creating components from a publicly accessible image or from a Dockerfile within your Github repository. In future we will support a wider variety of components.*

### Public image

If you want to create a component using a publicly accessible image hosted on an image registry (i.e. Docker Hub, GitHub Container Registry, Amazon Elastic Container Registry, ETC). To do this you’ll need to name your component, provide a URL to the public image and click the “Add component” button.

### Dockerfile from GitHub repository

To create a component from a Dockerfile within one of your GitHub repositories you’ll first need to connect your application to your GitHub account. Follow [this guide](./connect-to-github.md) if you’ve not done this yet. once you’ve connected your application to GitHub you can create a component using a Dockerfile from one of your repositories. From the Components page click the “Add component” button, enter a component name and select the “GitHub repository” under the Container Source option. You can now select a GitHub repository you’d like to build from, enter which branch to use and the directory the Dockerfile is located (defaults to the root directory). Once you’ve configured this information you can click the “Add component” button.

## Deploying your component

Now that you’ve created a component you can deploy it to any Installs you’ve created. To learn how to create an Install check out [this guide](./iam-role-for-installs.md). From the components page click the component you want to deploy and the component detail panel will open. From here click the “Deploy component” button to trigger a deployment of the component.
