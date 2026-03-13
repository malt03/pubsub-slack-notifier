gcloud config configurations activate fitty-prd
gcloud run deploy pubsub-slack-notifier \
  --source . \
  --platform managed \
  --region us-west1 \
  --no-allow-unauthenticated
