# Buildkit helm chart


Helps you deploy a buildkit cluster with mTLS support.


### Some notable decisions:
* Single instance of buildkit. As Buildkit isn't built to be horizontally scalable, we'll have to vertically scale it. 
* Envoy handles  mTLS for us & performs a TLS pass-through to buildkit. We haven't configured Buildkit to use mTLS. Couple benefits of this:
  * Future potential to allow scaling down Buildkit deployment to 0 when not in use & use Envoy's metrics to determine whether to scale up.
  * A rollout of newer version of Buildkit deployment can be done as Envoy would allow us to route traffic perfectly without downtime as it can hold the build requests until the new buildkit is ready.




### Setup of mTLS certs. 

- Generate certificates using:
```
SAN="example.com" docker buildx bake "https://github.com/moby/buildkit.git#master:examples/create-certs"
```
NOTE: buildx bake works on latest version of docker CLI and server. 


- Move .certs to certs
```  
mv .certs/ certs/     
```

- Create a secret with client & daemon certs:
```
kubectl -n buildkit create secret generic buildkit-secrets \
  --from-file=server.crt=certs/daemon/cert.pem \
  --from-file=server.key=certs/daemon/key.pem \
  --from-file=client.crt=certs/client/cert.pem \
  --from-file=client.key=certs/client/key.pem \
  --from-file=ca.crt=certs/daemon/ca.pem
```

- Create a secret for mTLS for envoy:
```
kubectl  -n buildkit create secret generic buildkit-envoy-certs \
  --from-file=ca.pem=certs/client/ca.pem \
  --from-file=server.pem=certs/daemon/cert.pem \
  --from-file=server-key.pem=certs/daemon/key.pem
```

- Create the remote driver on Docker:
```
docker buildx create \
  --name remote-kubernetes \
  --driver remote \
  --driver-opt cacert=${PWD}/certs/client/ca.pem,cert=${PWD}/certs/client/cert.pem,key=${PWD}/certs/client/key.pem \
  tcp://example.com:1234
```
NOTE: To test locally, we can edit `/etc/hosts` such that example.com resolves to BuildKit's address. Essentially, the SAN used during created should match with address. `tcp://<SAN>:<port>`.