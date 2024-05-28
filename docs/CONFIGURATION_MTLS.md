# Configuring mTLS: Guidelines and Instructions

**Step-1: Create a ca-config.cnf file**

```$xslt
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
req_extensions      = req_ext
x509_extensions     = v3_ca
[ req_distinguished_name ]
countryName                 = Country Name (2 letter code)
countryName_default         = US
stateOrProvinceName         = State or Province Name (full name)
stateOrProvinceName_default = New York
localityName                = Locality Name (eg, city)
localityName_default        = Albany
organizationName            = Organization Name (eg, company)
organizationName_default    = Kubviz
commonName                  = Common Name (e.g. server FQDN or YOUR name)
commonName_max              = 64
[ req_ext ]
subjectAltName = @alt_names
[ v3_ca ]
subjectAltName = @alt_names
[ alt_names ]
DNS.1 = kubviz-client-nats
DNS.2 = kubviz-client
DNS.3 = kubviz-agent
```

**Step-2: Create ca-cert.pem**

```bash
openssl genrsa -out ca-key.pem 4096
```

```bash
openssl req -new -x509 -days 365 -key ca-key.pem -out ca-cert.pem -subj "/C=US/ST=New York/L=Albany/O=Kubviz/CN=KubvizCA"
```

**Step-3: Create the Server Certificate**

```bash
openssl genrsa -out server-key.pem 4096
```

```bash
openssl req -new -key server-key.pem -out server-csr.pem -subj "/C=US/ST=New York/L=Albany/O=Kubviz/CN=kubviz-client-nats" -config ca-config.cnf -extensions req_ext
```

```bash
openssl x509 -req -days 365 -in server-csr.pem -CA ca-cert.pem -CAkey ca-key.pem -set_serial 01 -out server-cert.pem -extfile ca-config.cnf -extensions v3_ca
```

**Step-4: Create the Client Certificate**

```bash
openssl genrsa -out client-key.pem 4096
```

```bash
openssl req -new -key client-key.pem -out client-csr.pem -subj "/C=US/ST=New York/L=Albany/O=Kubviz/CN=kubviz-client" -config ca-congig.cnf -extensions req_ext
```

```bash
openssl x509 -req -days 365 -in client-csr.pem -CA ca-cert.pem -CAkey ca-key.pem -set_serial 02 -out client-cert.pem -extfile ca-config.cnf -extensions v3_ca
```

**step-5: Create the agent certificate**

```bash
openssl genrsa -out agent-key.pem 4096
```

```bash
openssl req -new -key agent-key.pem -out agent-csr.pem -subj "/C=US/ST=New York/L=Albany/O=Kubviz/CN=kubviz-agent" -config ca-config.cnf -extensions req_ext
```

```bash
openssl x509 -req -days 365 -in agent-csr.pem -CA ca-cert.pem -CAkey ca-key.pem -set_serial 02 -out agent-cert.pem -extfile ca-config.cnf -extensions v3_ca
```
**step-6: Create secrets**

```bash
kubectl create secret generic kubviz-client-ca-cert --from-file=client-cert.pem --from-file=client-key.pem --from-file=ca-cert.pem -n kubviz
```

```bash
kubectl create secret generic kubviz-agent-ca-cert --from-file=agent-cert.pem --from-file=agent-key.pem --from-file=ca-cert.pem -n kubviz
```

```bash
kubectl create secret generic kubviz-server-ca-cert --from-file=server-cert.pem --from-file=server-key.pem --from-file=ca-cert.pem -n kubviz
```

#### if you want to enable mtls add the secret name in client/values.yaml also mtls.enabled: true

**Step-7: Add the secret name in client/value.yaml**

Below is the nats configuration

```yaml
tls:
    secret:
      name: kubviz-server-ca-cert
    ca: "ca-cert.pem"
    cert: "server-cert.pem"
    key: "server-key.pem"
    verify: true
    verify_and_map: true
...
```

**Step-8: Add the secret name in client/value.yaml**

```yaml
mtls:
  enabled: true
  secret:
    name: kubviz-client-ca-cert
...
```

**Step-9: Add the secret name in agent/value.yaml**

```yaml
mtls:
  enabled: true
  secret:
    name: kubviz-agent-ca-cert
...
```
