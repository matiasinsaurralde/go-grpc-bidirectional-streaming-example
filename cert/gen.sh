rm -f *.pem *.srl *.cnf

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=US/ST=New York/L=New York/O=Test/OU=Test/CN=127.0.0.1/emailAddress=test@test.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=US/ST=New York/L=New York/O=Test/OU=Test/CN=127.0.0.1/emailAddress=test@test.com"

echo subjectAltName = IP:127.0.0.1 > extfile.cnf

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile extfile.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text
