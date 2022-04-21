# Go, Twitter and k8s Orchestration

Grabs latest tweets given a list of search terms and logs them out. Requires [Minikube](https://minikube.sigs.k8s.io/docs/start/) locally.
## Set up:

1. Create a [Twitter API key](https://developer.twitter.com/en/docs/twitter-api/getting-started/getting-access-to-the-twitter-api)
2. Run the following command to have your secrets set for Kubernetes
```
kubectl create secret generic twitter-keys \
  --from-literal=CLIENT_ID=<client id> \
  --from-literal=CLIENT_SECRET=<client secret>
```
3. Run `kubectl apply -f deploy.yaml`
4. Check logs with `kubectl logs orch-go -c app` and verify orchestration ran 
5. Profit 💸

![go](https://user-images.githubusercontent.com/535651/164535962-08e6eb67-ea61-45d2-830a-e2eec5406296.gif)