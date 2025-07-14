openssl genrsa -out ca-key.pem 4096

openssl req -x509 -new -nodes -sha512 -days 3650 \
	 -subj "/C=CN/ST=Russia/L=Russia/O=example/OU=Personal/CN=localhost" \
	 -key ca-key.pem \
	 -out ca-cert.pem


openssl genrsa -out server-key.pem 4096

openssl req -sha512 -new \
	    -subj "/C=CN/ST=Russia/L=Russia/O=example/OU=Personal/CN=localhost" \
	        -key server-key.pem \
		    -out server-cert.pem

cat > v3.ext <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1=localhost
DNS.2=127.0.0.1
EOF

openssl x509 -req -sha512 -days 3650 \
	-extfile v3.ext \
	-CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial \
	-in server-cert.pem \
	-out server-cert.pem


