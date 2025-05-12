Journal of what I do each day

# Set up docker and goose (May 6)
Finished setting up the `auth` database for the `auth` microservice.

`docker compose up`

`docker compose down -v`: `-v` to remove all volumes created.



# Implement auth (May 9)

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


# Some readings on consistency and PostgreSQL locks (May 10)

## PostgreSQL Lock and Transactions
[+] In Postgres, implicitly acquired lock of the INSERT query stays active until the end of the transaction.

[+] By *default*, PostgreSQL uses [**Read Committed Isolation Level**](https://www.postgresql.org/docs/7.1/xact-read-committed.html), which will NEVER read a `SELECT` query sees only data committed before the query began and never sees either uncommitted data or changes committed during query execution by concurrent transactions (except for the changes made by the transaction we're in). 

[+] For `UPDATE ... WHERE` clauses, whenever PostgreSQL finds a row that matches the filter, that row will be locked and updated. If locking a row is blocked by a concurrent query, PostgreSQL waits until the lock goes away. Then it re-evaluates the filter condition and either moves on (if the condition no longer applies on account of a concurrent modification) or it locks and updates the modified row.

[!] However, this means that two *successive* `SELECT`s can see different data, even though they are within a single transaction, when other transactions commit changes during execution of the first `SELECT`.

[!] **DEADLOCK**: "If you don't explicit lock a table using LOCK statement, it will be implicit locked only at the first UPDATE, INSERT, or DELETE operation. If you don't exclusive lock the table before the select, some other user may also read the selected data, and try and do their own update, causing a deadlock while you both wait for the other to release the select-induced shared lock so you can get an exclusive lock to do the update." See more [here](https://www.postgresql.org/docs/6.4/sql-lock.htm)

### Potential Deadlock Issue with Concurrent Transfers

Concurrent transactions involving insertions into the `transfers` table (which has foreign key constraints on the `accounts` table) followed by selections with `FOR UPDATE` on the `accounts` table can lead to deadlocks in PostgreSQL.

**Scenario:**

Consider two concurrent transactions (tx1 and tx2) executing the following sequence of operations:

1.  **tx1:** `INSERT INTO transfers(from_account_id, to_account_id, amount) VALUES (1, 2, 10) RETURNING *;`
2.  **tx2:** `INSERT INTO transfers(from_account_id, to_account_id, amount) VALUES (1, 2, 10) RETURNING *;`
3.  **tx2:** `SELECT * FROM accounts WHERE id = 1 FOR UPDATE;`
4.  **tx1:** `SELECT * FROM accounts WHERE id = 1 FOR UPDATE;`

**Explanation of the Potential Deadlock:**

* When each `INSERT` statement executes, PostgreSQL might acquire shared locks on the referenced rows in the `accounts` table (in this case, the row with `id = 1`) to verify the foreign key constraints.
* Subsequently, when both transactions attempt to acquire an exclusive lock on the same `accounts` row (`id = 1`) using `SELECT ... FOR UPDATE`, step 3 is waiting for the release of the shared lock from step 1, and step 4 is waiting for the release of the *same* shared lock from step 2.

**Solution**

Use `FOR **NO KEY** UPDATE**` instead of `FOR UPDATE` in the `SELECT` statement to tell PostgreSQL that we're NOT modifying the foreign key, so it doesn't acquire the shared lock. Of course, this solution only works if we're actually NOT modifying the foreign key ^_^


Read [this](https://blog.christianposta.com/microservices/the-hardest-part-about-microservices-data/).

**Transactional Boundaries**: the smallest unit of **atomicity** that you need with respect to the business invariants.

For transactions that span multiple services, we should keep the **transactional boundaries** (between microservices) as small as possible. That is, we split, say a *transfer* global transaction into *individual transactions*, one for each microservice. This ensures **scalability**. 

But how do we ensure **consistency**? 

=>  Between transaction boundaries and between bounded contexts, use **events** to communicate consistency. Events are **immutable** structures that capture an interesting point in time that should be broadcast to peers (microservices). Peers will listen to the events in which they’re interested and make decisions based on that data, store that data, store some derivative of that data, update their own data based on some decision made with that data, etc, etc.

=> use an ACID database and stream changes to that database to a persistent, replicated log like **Apache Kafka** using something like Debezium and deduce the events using some kind of event processor/steam processor.



# Implemented the account service. Decided that the close future plan is to implement orchestration-based SAGA with gRPC first before adding Kafka. (May 11)

Some readings:

1. A very comprehensive explanation of orchestration vs choreography SAGAs: https://livebook.manning.com/book/microservices-patterns/chapter-4/143

## Orchestration-based SAGA
**Orchestration SAGA** is **NOT** ACID. It's only ACD: That’s because the updates made by each of a saga’s local transactions are immediately visible to other sagas once that transaction commits. This behavior can cause two problems. First, other sagas can change the data accessed by the saga while it’s executing. And other sagas can read its data before the saga has completed its updates, and consequently can be exposed to inconsistent data.

This lack of isolation potentially causes what the database literature calls **anomalies**. An anomaly is when a transaction reads or writes data in a way that it wouldn’t if transactions were executed one at time. 3 anomalies include:

1. Lost updates— One saga overwrites without reading changes made by another saga.
2. Dirty reads— A transaction or a saga reads the updates made by a saga that has not yet completed those updates.
3. Fuzzy/nonrepeatable reads— Two different steps of a saga read the same data and get different results because another saga has made updates.


But then [again](https://blog.christianposta.com/microservices/the-hardest-part-about-microservices-data/), one thing that we should understand: distributed systems are finicky. There are very few guarantees if any we can make about anything in a distributed system in bounded time (things WILL fail, things are non-deterministically slow or appear to have failed, systems have non-synchronized time boundaries, etc), so why try to fight it? What if we embrace this and bake it into our consistency models across our domain? What if we say “between our necessary transactional boundaries we can live with other parts of our data and domain to be reconciled and made consistent at some later point in time”?

However, there are COUNTERMEASURES to this lack of isolation:

1. **COMMUTATIVE** updates: a money transfer = a credit to 1 account AND a debit to another account. => For failed transferes, we reverse the operation (debit to the 1st account and credit the 2nd account). The *reversed* operations are called **compensating transactions**.
2. **Semantic locks**:  sets a flag in any record that it creates or updates. The flag indicates that the record isn’t *committed* and could potentially change. Still need to read more on this...



gRPC style: https://protobuf.dev/programming-guides/style/

Very good series on gRPC: https://www.youtube.com/watch?v=YzypniHHU3w&list=PLy_6D98if3UJd5hxWNfAqKMr15HZqFnqf&index=6

Data dependencies among microservices such as creating an account requires a user id might make it hard to maintain the context boundaries between microservices when an external API of a microservice is updated. Though, it is unlikely if the only thing we're passing in gRPC calls are IDs, which are all uuid.UUID. This is  why I'm also having second thoughts about merging the **account** service and **transaction** service all together.

Right now, implementing orchestration-based is simpler. In the future, I'm interested in implementing a choreography-based using Kafka :)


# Decided to merge the Account and Transaction services together since they are tightly coupled (May 12)

