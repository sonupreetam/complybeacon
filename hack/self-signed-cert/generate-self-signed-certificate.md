> [!WARNING]
> This is self-signed certificate, only use for testing/development.
> Generate all certificate files under [self-signed-cert dir](../self-signed-cert)

### 1. Create the new Root CA key

```shell
openssl genrsa -out my-ca.key 2048
```

### 2. Create the new Root CA certificate

```shell
openssl req -x509 -new -nodes -key my-ca.key -sha256 -days 365 \
-subj "/CN=ComplyBeacon Root CA" \
-extensions v3_ca -config openssl.cnf \
-out my-ca.crt
```

### 3. Create the server's private key
```shell
openssl genrsa -out compass.key 2048
```

### 4. Create a Certificate Signing Request (CSR) for the server
```shell
openssl req -new -key compass.key -out compass.csr -config openssl.cnf
```

### 5. Use your new Root CA to sign the server's CSR
```shell
openssl x509 -req -in compass.csr -CA my-ca.crt -CAkey my-ca.key -CAcreateserial \
-out compass.crt -days 365 -sha256 \
-extfile openssl.cnf -extensions v3_req
```
