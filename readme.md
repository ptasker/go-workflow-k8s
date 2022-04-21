# Go and k8s Orchestration

## Set up:

1. Create a Twitter API key
2. Run the following command to have your secrets set for Kubernetes
```
kubectl create secret generic twitter-keys \
  --from-literal=CLIENT_ID=<client id> \
  --from-literal=CLIENT_SECRET=<client secret>
```
3. Run `kubectl apply -f deploy.yaml`
4. Profit