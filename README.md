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
