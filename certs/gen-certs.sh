# Generate server key and certificate with SAN
openssl genrsa -out tls.key 2048
openssl req -new -key tls.key \
  -subj "/CN=${SERVICE}.${NAMESPACE}.svc" \
  -out tls.csr

# Create SAN config
cat > san.cnf <<EOF
[req]
req_extensions = v3_req
[v3_req]
subjectAltName = DNS:${SERVICE}.${NAMESPACE}.svc,DNS:${SERVICE}.${NAMESPACE}.svc.cluster.local
EOF

openssl x509 -req -days 365 \
  -in tls.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -extensions v3_req \
  -extfile san.cnf \
  -out tls.crt