Journal of what I do each day

# May 6 - Set up docker and goose

Finished setting up the `auth` database for the `auth` microservice.

`docker compose up`

`docker compose down -v`: `-v` to remove all volumes created.

# May 9 - Implement auth

Tasks done:

1. Refactored code to follow the Controller-Service-Repository [philosophy](https://medium.com/@shershnev/layered-architecture-implementation-in-golang-6318a72c1e10).
2. Implement database transactions for creating a user.
3. Finished `repository` for `auth`.
4. Create integration tests for step 2 by spinning up the auth PostgreSQL docker container.

Some readings done:

1. Follow the philosophy Controller-Service-Repository: <https://medium.com/@shershnev/layered-architecture-implementation-in-golang-6318a72c1e10>
2. Scaling Kubernetes using Go: <https://nyadgar.com/posts/scaling-grpc-with-kubernetes-using-go/>

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

# May 10 - Some readings on consistency and PostgreSQL locks

## PostgreSQL Lock and Transactions

[+] Postgres isolation levels: <https://www.postgresql.org/docs/current/transaction-iso.html>.

[+] In Postgres, implicitly acquired lock of the INSERT query stays active until the end of the transaction.

### Read-Committed (default)

[+] By *default*, PostgreSQL uses [**Read Committed Isolation Level**](https://www.postgresql.org/docs/7.1/xact-read-committed.html), which will NEVER read **uncommitted** changes. A `SELECT` query sees only data committed before the query began and never sees either uncommitted data or changes committed during query execution by concurrent transactions (except for the changes made by the transaction we're in).

[+] For `UPDATE ... WHERE` clauses, whenever PostgreSQL finds a row that matches the filter, that row will be locked and updated. If locking a row is blocked by a concurrent query, PostgreSQL waits until the lock goes away. Then it re-evaluates the filter condition and either moves on (if the condition no longer applies on account of a concurrent modification) or it locks and updates the modified row.

[!] However, this means that two *successive* `SELECT`s can see different data, even though they are within a single transaction, when other transactions commit changes during execution of the first `SELECT`.

[!] **DEADLOCK**: "If you don't explicit lock a table using LOCK statement, it will be implicit locked only at the first UPDATE, INSERT, or DELETE operation. If you don't exclusive lock the table before the select, some other user may also read the selected data, and try and do their own update, causing a deadlock while you both wait for the other to release the select-induced shared lock so you can get an exclusive lock to do the update." See more [here](https://www.postgresql.org/docs/6.4/sql-lock.htm)

Because Read Committed mode starts *each* command with a *new snapshot* that includes all transactions committed up to that instant, subsequent commands in the same transaction will see the effects of the committed concurrent transaction in any case. The point at issue above is whether or not a single command sees an absolutely consistent view of the database.

### Repeatable-Read

<https://www.postgresql.org/docs/current/transaction-iso.html#XACT-REPEATABLE-READ>

This level is different from Read Committed in that a query in a repeatable read transaction sees a snapshot as of the start of the first non-transaction-control statement in the transaction, not as of the start of the current statement within the transaction. Thus, successive SELECT commands within a single transaction see the same data, i.e., they do not see changes made by other transactions that committed after their own transaction started.

UPDATE, DELETE, MERGE, SELECT FOR UPDATE, and SELECT FOR SHARE commands behave the same as SELECT in terms of searching for target rows: they will only find target rows that were committed as of the transaction start time. However, such a target row might have already been updated (or deleted or locked) by another concurrent transaction by the time it is found. In this case, the repeatable read transaction will wait for the first updating transaction to commit or roll back (if it is still in progress). If the first updater rolls back, then its effects are negated and the repeatable read transaction can proceed with updating the originally found row. But if the first updater commits (and actually updated or deleted the row, not just locked it) then the repeatable read transaction will be rolled back with the message

`ERROR:  could not serialize access due to concurrent update`

Note that only updating transactions might need to be retried; read-only transactions will never have serialization conflicts.

### Potential Deadlock Issue with Concurrent Transfers

Concurrent transactions involving insertions into the `transfers` table (which has foreign key constraints on the `accounts` table) followed by selections with `FOR UPDATE` on the `accounts` table can lead to deadlocks in PostgreSQL.

**Scenario:**

Consider two concurrent transactions (tx1 and tx2) executing the following sequence of operations:

1. **tx1:** `INSERT INTO transfers(from_account_id, to_account_id, amount) VALUES (1, 2, 10) RETURNING *;`
2. **tx2:** `INSERT INTO transfers(from_account_id, to_account_id, amount) VALUES (1, 2, 10) RETURNING *;`
3. **tx2:** `SELECT * FROM accounts WHERE id = 1 FOR UPDATE;`
4. **tx1:** `SELECT * FROM accounts WHERE id = 1 FOR UPDATE;`

**Explanation of the Potential Deadlock:**

- When each `INSERT` statement executes, PostgreSQL might acquire shared locks on the referenced rows in the `accounts` table (in this case, the row with `id = 1`) to verify the foreign key constraints.
- Subsequently, when both transactions attempt to acquire an exclusive lock on the same `accounts` row (`id = 1`) using `SELECT ... FOR UPDATE`, step 3 is waiting for the release of the shared lock from step 1, and step 4 is waiting for the release of the *same* shared lock from step 2.

**Solution**

Use `FOR **NO KEY** UPDATE**` instead of `FOR UPDATE` in the `SELECT` statement to tell PostgreSQL that we're NOT modifying the foreign key, so it doesn't acquire the shared lock. Of course, this solution only works if we're actually NOT modifying the foreign key ^_^

Read [this](https://blog.christianposta.com/microservices/the-hardest-part-about-microservices-data/).

**Transactional Boundaries**: the smallest unit of **atomicity** that you need with respect to the business invariants.

For transactions that span multiple services, we should keep the **transactional boundaries** (between microservices) as small as possible. That is, we split, say a *transfer* global transaction into *individual transactions*, one for each microservice. This ensures **scalability**.

But how do we ensure **consistency**?

=>  Between transaction boundaries and between bounded contexts, use **events** to communicate consistency. Events are **immutable** structures that capture an interesting point in time that should be broadcast to peers (microservices). Peers will listen to the events in which they’re interested and make decisions based on that data, store that data, store some derivative of that data, update their own data based on some decision made with that data, etc, etc.

=> use an ACID database and stream changes to that database to a persistent, replicated log like **Apache Kafka** using something like Debezium and deduce the events using some kind of event processor/steam processor.

# May 11 - Implemented the account service. Decided that the close future plan is to implement orchestration-based SAGA with gRPC first before adding Kafka

Some readings:

1. A very comprehensive explanation of orchestration vs choreography SAGAs: <https://livebook.manning.com/book/microservices-patterns/chapter-4/143>

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

gRPC style: <https://protobuf.dev/programming-guides/style/>

Very good series on gRPC: <https://www.youtube.com/watch?v=YzypniHHU3w&list=PLy_6D98if3UJd5hxWNfAqKMr15HZqFnqf&index=6>

Data dependencies among microservices such as creating an account requires a user id might make it hard to maintain the context boundaries between microservices when an external API of a microservice is updated. Though, it is unlikely if the only thing we're passing in gRPC calls are IDs, which are all uuid.UUID. This is  why I'm also having second thoughts about merging the **account** service and **transaction** service all together.

Right now, implementing orchestration-based is simpler. In the future, I'm interested in implementing a choreography-based using Kafka :)

# May 12 - Decided to merge the Account and Transaction services together since they are tightly coupled

## PostgreSQL concurrent statements in the SAME transaction

If multiple concurrent statements are running within the SAME transaction, PostgreSQL is gonna throw an error like:

```bash
Error:       Received unexpected error:
                         pq: unexpected Parse response 'C'
```

This is because we're trying to start a new statement before reading all of the rows of the preceding statement.

In practice, this should never happen as developers shouldn't run concurrent statements within the same transaction. What they mean to do is run multiple concurrent TRANSACTIONs, NOT STATEMENTs.

## PostgreSQL's `INSERT` will raise an error if multiple insertions have the same unique index (such as the primary key)

INSERT into tables that lack unique indexes will not be blocked by concurrent activity. Tables with unique indexes might block if concurrent sessions perform actions that lock or modify rows matching the unique index values being inserted; the details are covered in Section 62.5. ON CONFLICT can be used to specify an alternative action to raising a unique constraint or exclusion constraint violation error. (See ON CONFLICT Clause [below](https://www.postgresql.org/docs/current/sql-insert.html).)

# May 13 - Refactored code for the repository layers to leave the transaction logic to the service layer

## `protoc`-generated files

In summary, the `.pb.go` file deals with message definitions, while the `_grpc.pb.go` file deals with service/RPC definitions. This separation allows you to use the message definitions without requiring the gRPC dependencies if needed.

# May 20 - Added the service layer for `auth` microservice

# May 23 - Finished testing and idempotency checking for the service layer of the `account` microservice. Decided that only the ReadCommitted Isolation Level is necessary

We can use LevelReadCommitted for CreateAccount because multiple concurrent requests creating different accounts don't affect each other in any way. Multiple concurrent requests creating the *same* acocunt will wait for each other when they try to claim the idempotency key, then the key's status will be checked to make sure the same request is not executed twice.

Can we use the same isolation level for CreateTransaction? The argument is the same for multiple concrrent requests with the same idempotency key. But, for multiple concurrent requests (with different idempotency keys) updating the account balance of the **same account**, there has to be a serial order of update. As per the [Read-Committed Isolation Level by PostgreSQL](https://www.postgresql.org/docs/current/transaction-iso.html#XACT-READ-COMMITTED), the 2nd conflicting update will apply the same operation to the **updated** row. Luckily, the way we update the account balance is: `SET balance = balance + $2`, so the 2nd update will use the correct updated value from the 1st update.

# May 27 - Read about authentication methods: JWT vs Traditional Session


Read [OWASP Authentication Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#consider-strong-transaction-authentication)

If decided to use JWT, consult the [OWASP JWT Cheatsheet](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.md#token-sidejacking)

Someone followed the OWASP guidelines and implement JWT [here](https://hasura.io/blog/best-practices-of-using-jwt-with-graphql#silent-refresh).


To summarise:

1. Use a strong hashing algorithm.
2. Make sure the JWT library doesn't accept `none` as the hashing algo in your JWT token.
3. To protect against a stolen JWT access token (JWT token-sidejacking), use a <u>fingerprint</u>:
    - Server generates a random string, which will be sent to the client as a *hardened* cookie, with  HttpOnly + Secure, SameSite, Max-Age, and cookie prefixes.
    - Then, the SHA256 of that random string is embedded within the JWT token.
    - To validate: 
        + Compute the fingerprint in the request against the hashed fingerprint in the JWT token.
4. A "logout" of a user - the JWT token expires sooner than its `expired_at` time - can be simulated by deleting the JWT from the client storage. Deleting can happen if we store the JWT token in `sessionStorage`, since it will be cleared automatically when the user closes the browser. 

## **sessionStorage** vs **localStorage**
*sessionStorage* is per-tab, while *localStorage* persists across tabs and browser restarts. Hence, *localStorage* is preferred for usabilitiy.

As per the [OWASP guidelines](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.md#token-storage-on-client-side), JWT Tokens in *any storage* should:
    - Have short expiration times (e.g. 15-30 minutes idle timeout, 8-hour absolute timeout).
    - Implement `refresh_tokens`.

But isn't having `refresh_tokens`` the same as a traditionally stateful session? Well, kinda, but it reduces the amount of queries you need for authentication:
    - In stateful applications, you will have to pass the request to the `auth` service and let it query its database every time.
    - However, with JWTs, *any* microservice can validate the JWT. Hence, the only "state" that we need to query from the `auth` service is during *revalidating/refreshing* an access token. 

## So we use *sessionStorage*. What will happen if I am logged in on different tabs?
One way of solving this is by introducing a global event listener on localStorage. Whenever we remove this logout key in localStorage on one tab, the listener will fire on the other tabs and trigger a "logout" too and redirect users to the login screen.





# May 28 - Implement the service layer and gRPC proto and install buf.build for the `auth` microservice


Use [protovalidate](https://buf.build/docs/protovalidate/quickstart/grpc-go/) to enforce rules on gRPC messages.

[buf.yaml](./auth/buf.yaml) stores where to look for `.proto` input files, and the buf.build modules we need to generate code. 

[buf.gen.yaml](./auth/buf.gen.yaml) stores the locations of output files. Use [managed mode](https://buf.build/docs/generate/managed-mode/#go) to override the default output "location pattern".


# May 29 - Implement JWT validation at the API Gateway microservice.

Installed the [chi]() library for http router.

Implement JWT authentication middleware, and form parsing middleware with the following [considerations](https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body).


# May 30 

JWT token validation should happen at the API Gateway. Then the API Gateway would pass the userID of the validated token down to other microservices. 

In the case of renewing an access token, the API Gateway would need to pass both the userID of the validated JWT and the `refresh_token` cookie to the `auth` microservice. This allows the `auth` microservice to query the DB for the `refresh_token`, and check if its userID matches the one from the validated JWT to prevent a malicious attacker who has stolen a JWT token from gaining a new access token to the victim account.



