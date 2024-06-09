# Server Beta
Second backend for group project

## Pipelines
![Tests](https://github.com/wwi21seb-projekt/server-beta/actions/workflows/ci.yml/badge.svg?branch=main&event=push)\
![Deployment](https://github.com/wwi21seb-projekt/server-beta/actions/workflows/cd.yml/badge.svg?branch=main&event=push)

# Prerequisites
- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go (Version 1.21.4)](https://go.dev/)

## Usage
````
git clone https://github.com/wwi21seb-projekt/server-beta.git
cd server-beta
make all
./bin/server-beta -port 8080
````

If no port is specified, the server will run on the default port `:8080`.

The user needs to have PostgreSQL installed and running. Additionally, the user needs to create an `.env` file with the following information:

| Variable          | Description                                                                      |
|-------------------|----------------------------------------------------------------------------------|
| JWT_SECRET        | Secret key for JSON Web Token (JWT) authentication                               |
| DB_HOST           | Hostname or IP address of the PostgreSQL database server                         |
| DB_PORT           | Port number of the PostgreSQL database server                                    |
| DB_SSL_MODE       | SSL mode for the database connection (e.g., disable, require, etc.)              |
| DB_NAME           | Name of the PostgreSQL database                                                  |
| DB_USER           | Username for the PostgreSQL database                                             |
| DB_PASSWORD       | Password for the PostgreSQL database                                             |
| PROXY_HOST        | Hostname or IP address of the proxy server                                       |
| SERVER_URL        | URL of the server                                                                |
| EMAIL_HOST        | Hostname or IP address of the email server                                       |
| EMAIL_PORT        | Port number of the email server                                                  |
| EMAIL_ADDRESS     | Email address used for sending emails                                            |
| EMAIL_PASSWORD    | Password for the email address                                                   |
| VAPID_PRIVATE_KEY | VAPID private key for web push notifications                                     |
| VAPID_PUBLIC_KEY  | VAPID public key for web push notifications                                      |
| IMAGES_PATH       | Path to the directory where images are stored                                    |
| GIN_MODE          | Mode of the application (e.g., debug, release)                                   |


An example `.env` file (`.env.example`) can be found in the root directory of the project.