Journal of what I do each day

# Set up docker and goose (May 6)
Finished setting up the `auth` database for the `auth` microservice.

`docker compose up`

`docker compose down -v`: `-v` to remove all volumes created.



# Setup gRPC for auth (May 9)

Tasks done:

1. Refactored code to follow the Controller-Service-Repository [philosophy](https://medium.com/@shershnev/layered-architecture-implementation-in-golang-6318a72c1e10).
2. Implement database transactions for creating a user.
3. Finished `repository` for `auth`.
4. Create integration tests for step 2 by spinning up the auth PostgreSQL docker container.
 

Some readings done:

1. Follow the philosophy Controller-Service-Repository: https://medium.com/@shershnev/layered-architecture-implementation-in-golang-6318a72c1e10
2. Scaling Kubernetes using Go: https://nyadgar.com/posts/scaling-grpc-with-kubernetes-using-go/

In Golang, the Controller is called the Handler instead.

So:

- Handler will create and receive gRPC calls from the API Gateway.
- Service will hanlde all the business logic (e.g. user is authorized?)
- Repository will handle all the connections to the database.


Directory sturcture:

```bash
auth-service/
├── cmd/
│   └── server/              # Entry point (main.go)
│       └── main.go
│
├── internal/                # Private app logic (not exported)
│   ├── handler/             # gRPC handler (implements proto interface)
│   │   └── user_handler.go
│   │
│   ├── service/             # Business logic
│   │   └── user_service.go
│   │
│   ├── repository/          # Database access
│   │   └── user_repository.go
│   │
│   ├── model/               # Entity definitions for gRPC 
│   │   └── user.go
│   │
│   ├── initializer/         # Init logic: DB, config, env
│   │   ├── db.go
│   │   ├── env.go
│   │   └── logger.go
│   │
│   └── util/                # Helpers (e.g., password hashing)
│       └── password.go
│
├── proto/                   # gRPC .proto files
│   ├── auth.proto
│   └── gen/                 # Generated Go code (protoc output)
│       └── auth.pb.go
│
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile                 # (Optional) build / run commands
└── README.md
```


## SQL transactions with `sqlc`

`sqlc` only supports individual queries. To extend it to transactions (multiple queries), we need to use `DB` from `database/sql` to create transactions.


We can do this by creating a new struct:

```Go
type AuthRepository struct {
	queries *sqlc.Queries
	db *sqlx.DB
}
```

Then we can use it like:

```Go
// wrap a function in a transaction and execute it
func (r *AuthRepository) execTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil) // create a transaction
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := r.queries.WithTx(tx) // return a new queries object with the transaction

	if err := fn(q); err != nil { // execute the function with the transaction
		return err
	}

	return tx.Commit()
}

// create user in a transaction: check if user already exists, if not create user
func (r *AuthRepository) CreateUserTx(ctx context.Context, user *model.User) (*sqlc.User, error) {

	var createdUser sqlc.User

	err := r.execTx(ctx, func(q *sqlc.Queries) error {
		var err error

		// check if user already exists
		_, err = q.GetUserByEmail(ctx, user.Email)
		...
        ....
		// create user
		createdUser, err = q.CreateUser(ctx, sqlc.CreateUserParams{
			ID:           uuid.New(),
			Email:        user.Email,
			PasswordHash: passwordHash,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return &createdUser, err

}
```

