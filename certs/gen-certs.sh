#!/bin/bash
SERVICE=webhook-service
NAMESPACE=default
SECRET=webhook-tls

# Generate CA key and certificate
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -subj "/CN=webhook-ca" -out ca.crt


openssl genrsa -out tls.key 2048
openssl req -new -key tls.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out tls.csr
openssl x509 -req -days 365 -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt



kubectl create secret generic ${SECRET} \
  --from-file=tls.crt=tls.crt \
  --from-file=tls.key=tls.key \
  --namespace=${NAMESPACE}


echo "CA Bundle (use this in webhook configuration):"
cat ca.crt | base64 | tr -d '\n'
echo ""