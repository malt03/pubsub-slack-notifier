## deploy

```
gcloud run deploy pubsub-slack-notifier \
  --source . \
  --platform managed \
  --region $REGION \
  --no-allow-unauthenticated
```
