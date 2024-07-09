````
server-beta/
├── .github/workflows/
│   ├── ...
├── cmd/
│   ├── server-beta/
│       ├── main.go
├── internal/
│   ├── controllers/
│       ├── user_controller.go
│       ├── user_controller_test.go
│       ├── ...
│   ├── customerrors/
│       ├── errors.go
│       ├── model.go
│   ├── initializers/
│       ├── connect_to_db.go
│       ├── ...
│   ├── middleware/
│       ├── ...
│   ├── models/
│       ├── user.go
│       ├── ...
│   ├── repositories/
│       ├── user_repository.go
│       ├── user_repository_mock.go
│       ├── ...
│   ├── router/
│       ├── router.go
│   ├── routines/
│       ├── ...
│   ├── services/
│       ├── user_service.go
│       ├── ...
│   ├── utils/
│       ├── validator.go
│       ├── ...
├── Makefile
├── README.md
├── ...
````
