# k8s-heartbeat

A HTTP api that exposes monitoring endpoints to check the availability of deployments running in a kubernetes cluster.

## Why

This project was designed out of the need to actively monitor deployments running in a kubernetes cluster. 
A remote status page or third-party application can then retrieve the status by pooling the endpoint related to a specific deployment, as we have done with the NixStats status page for example. 

This type of active monitoring does not require that you - the developer or cluster administrator - have to implement this logic manually sometimes having to develop some weird email integration or provide the storage to keep track of which service is down and whatnot.

If the deployment, or actually, the pods running for that deployment are degraded by some sort of outage the endpoint will automatically reflect that and when the service resumes it's normal execution the next request will succeed and the upstream monitoring tool can automagically close the incident for you.

## Usage

> Please ensure that your deployments have a livenessProbe set up so the monitoring will be more precise.

A kubernetes deployment configuration to deploy this project can be found in the config directory.

You can use an ingress or cluster ip to expose the deployment, after that you should be able to query the status with an HTTP GET or HEAD request:

```bash
curl -XHEAD https://status.example.com/api/healthz/kube-system/deployment/kube-proxy?token=dGVzdDp0ZXN0
```

In order to prevent malicious actors from disclosing private data about your cluster a Basic auth and rate limiting middleware are implemented, please check the Environment variables available below.

## Endpoints

The server has the following endpoints:

### GET or HEAD /health - The server's own health checking

`curl http://localhost:8080/healthz`

### GET or HEAD /api/healthz/{namespace}/deployment/{component}?token=xxx - Health check for a given deployment

`curl http://localhost:8080/api/healthz/kube-system/deployment/kube-proxy?token=dGVzdDp0ZXN0`


## Environment variables

```bash
PORT=8080 # the server's port
LOG_LEVEL=INFO 	# The log level for messages.
RATE_LIMIT=3600-H # Rate limiting, 3600 requests per hour
AUTH_TOKEN_BASIC= # Auth token for authorization, either send by the client via a "token" query param or Authorization Basic header. The server just compares the values, you may use base64 encoding if you wish and usinng HTTPS is highly recommended.
KUBECONFIG # Path to the kubernetes cluster config file, leave empty for in-cluster autodiscovery
```
