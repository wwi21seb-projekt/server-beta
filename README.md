# Server Beta
Second backend for group project

## Pipelines
![Tests](https://github.com/wwi21seb-projekt/server-beta/actions/workflows/ci.yml/badge.svg?branch=main&event=push)\
![Deployment](https://github.com/wwi21seb-projekt/server-beta/actions/workflows/cd.yml/badge.svg?branch=main&event=push)

# Prerequisites
Make sure you have the following tools and services installed, configured and running on your system:

- [Git](https://git-scm.com/)
````cmd
    apt-get install git
````
- [Make](https://www.gnu.org/software/make/)
````cmd
    apt-get install make
````
- [Go (Version 1.21.4)](https://go.dev/)
````cmd
    wget https://golang.org/dl/go1.21.4.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
````
- [PostgreSQL](https://www.postgresql.org/)
````cmd
    apt-get install postgresql
    -u postgres psql
````
````sql
    CREATE DATABASE serverbetadb;
    CREATE USER goserveruser WITH PASSWORD 'password';
    GRANT ALL PRIVILEGES ON DATABASE serverbetadb TO goserveruser;
````
The necessary tables are created automatically by the server when it is started for the first time.
- [nginx](https://www.nginx.com/) (optional)
````cmd
    apt-get install nginx
````

To use nginx as a reverse proxy, change the configuration file in `/etc/nginx/sites-available/default` to the configuration specified in the `nginx.conf` file in the root directory of the project. Restart the nginx service with `sudo systemctl restart nginx` after changing the configuration.

With a nginx reverse proxy, the server can be accessed via the domain name or IP address of the server. Additionally, with e.g. Let's Encrypt, a free SSL certificate can be obtained to secure the connection to the server.
````cmd
    apt install certbot
    apt install python3-certbot-nginx
    certbot --nginx
````
## Usage
Clone the repository, build the project and run the server with the following commands:
````
git clone https://github.com/wwi21seb-projekt/server-beta.git
cd server-beta
make all
./bin/server-beta -port 8080
````

If no port is specified, the server will run on the default port `:8080`.

Additionally, the user needs to create an `.env` file with the following information. An example `.env` file (`.env.example`) can be found in the root directory of the project.

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
| GIN_MODE          | Mode of the application (e.g., debug, release)                                   |

In deployment, a systemctl service can be created to run the server as a service. The following steps are necessary to create a service:

1. Create a new service file in `/etc/systemd/system/server-beta.service` with the content specified in the `server-beta.service` file in the root directory of the project.
2. Reload the systemctl daemon.
````cmd
    systemctl daemon-reload
````
3. To automatically start the server on boot, enable the service.
````cmd
    systemctl enable server-beta
````
4. Start the server.
````cmd
    systemctl start server-beta
````
5. Check the status of the server.
````cmd
    systemctl status server-beta
````

The logs of the server can be viewed using the `journalctl` commands.
````cmd
    journalctl -u server-beta
````

## CI/CD-Pipeline
The project uses GitHub Actions for continuous integration and continuous deployment. The workflows are specified in the `.yaml`-files in the `.github/workflows`-directory.

To use the CD-Pipeline in GitHub Actions, SSH access to the server, where the project shall be deployed, must be configured. 
In order for the pipeline to work, several secrets in the GitHub Repository settings. The following secrets are needed:

| Variable           | Description                                                       |
|--------------------|-------------------------------------------------------------------|
| SERVER_APP_PATH    | File path where the built executable file is placed on the server |
| SERVER_HOST        | IP address of the server                                          |
| SERVER_SSH_PORT    | Open SSH Port of the server                                       |
| SERVER_SSH_PRIVATE | Private SSH key to access the server                              |
| SERVER_SSH_Public  | Public SSH key to access the server                               |
| SERVER_USER        | Username of the user that accesses the server using SSH           |


