# Add a Component

In Nuon, a component can be one of the following:

-   a Docker image that gets built from a Docker file in a GitHub repository, e.g., [httpbin](https://github.com/postmanlabs/httpbin)
-   a Docker image that's in Docker hub or your ECR repo
-   a Helm chart, either from your private repo or publicly accessible, e.g.,[cp-helm-charts](https://github.com/confluentinc/cp-helm-charts)

To add a component:

1.  Click **Add a component**, from the Overview page or Components page.
2.  Enter the name of the component, in the panel that appears.
3.  Configure the source, following the instructions in [this section](configure-source.md).
4.  Configure the deployment, following the instructions in [this section](configure-deployment.md).
5.  Click **Add a component**, at the bottom of the panel.

![image](images/add-component.png)

The component will now be displayed on the Components page. 

**Note:** Currently, Nuon only supports creating components from a publicly accessible Docker image, or from a Dockerfile in your GitHub repository.