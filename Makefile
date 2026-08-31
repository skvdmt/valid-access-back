proto-gen:
	protoc -I api ./api/valid-access.proto --go_out=./gen/go --go-grpc_out=require_unimplemented_servers=false:./gen/go

certs-gen-auth:
	openssl req -x509 -out ./certs/valid-access-back/cert.crt -keyout ./certs/valid-access-back/private.key \
  	-newkey rsa:2048 -nodes -sha256 \
  	-subj '/CN=localhost' -extensions EXT -config <( \
   	printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
