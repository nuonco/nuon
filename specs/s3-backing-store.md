# Feature Spec

We work with many large files such as: terraform plans, kube-manifests and more. Right now, we store them in postgres, 
but want to build a layer to store them in blob storage (and encrypted). Some of the problems we have right now is that 
fetching these files have to be loaded into memory, require much space in the database and are generally hard to work 
with.

We want to create a gorm type that will allow us to store large payloads in s3, and then easily retrieve them in 
different contexts. We want things such as the runner writing a plan or the ui fetching a plan to stream through a 
general blob endpoint. When an object is created, we should be able to return a blob-id and then work against that blob 
id using the ctl-api.

It's important this is easy to opt in, so for any object that we want to support blobs for, we will just append the 
`Blob` suffix.

## Implementation Plan

1. Build new gorm type that enables us to have a custom type embedded for easy accessing.
1. Build a data-converter that will automatically load files from s3, so they don't hit the temporal history.
1. Design the API for dealing with blobs.
1. Create an endpoint to upload blob objects, and fetch blob objects.

From there, we should start implementing this workflow across the product. We'll start by rolling it out in the plan 
layer, and then in the runner. That can come later, but for the rollout we just want this to start working. 
