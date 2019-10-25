# About doorman

This doorman micro-service was created for educational purposes and should not be used for production in the current state.

## Environment variables used in the container
`DATABASE_USERNAME` - MySQL database username  
`DATABASE_PASSWORD` - MySQL database password  
`DATABASE_HOST` - The host on which the database lives  
`DATABASE` - Name of the application database  

## Secrets
The doorman microservice expects two secrets - doorman-db and doorman-certs. With the former containing the database connection details while the latter contains the RSA private key and certificate file that will be used to sign and verify the JWT(JSON Web Token) token.

The container expects a TLS secret to be mounted at `/etc/doorman/certs` as a volume. The certificate should be of type RSA.

You can create the certificate in kubernetes as shown below:

1. Create the certificate secret

`kubectl create secret tls doorman-certs --cert=tls.crt --key=tls.key`

2. Create the database credentials secret

`kubectl create secret generic doorman-db --from-literal database=<db-name> --from-literal hostname=<hostname/ip> --from-literal password=<database-password> --from-literal username=<db-user>`

## Database
The doorman microservice uses MySQL to store user credentials and hashed passwords.

The database structure is shown below:

```
CREATE TABLE `user` (
  `created_at` varchar(20) DEFAULT NULL,
  `updated_at` varchar(20) DEFAULT NULL,
  `deleted_at` varchar(20) DEFAULT NULL,
  `id` varchar(48) NOT NULL,
  `firstname` varchar(48) DEFAULT NULL,
  `lastname` varchar(48) DEFAULT NULL,
  `username` varchar(48) NOT NULL,
  `password` varchar(128) DEFAULT NULL,
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB;
```
