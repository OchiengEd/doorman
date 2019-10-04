## Environment variables
`DATABASE_USERNAME` - MySQL database username  
`DATABASE_PASSWORD` - MySQL database password  
`DATABASE_HOST` - The host on which the database lives  
`DATABASE` - Name of the application database  

## Secrets
The container expects a TLS secret to be mounted at `/etc/doorman/certs` as a volume. The certificate should be of type RSA.
