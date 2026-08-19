# Gator

**Gator** is a lightweight command-line RSS feed aggregator and reader built in Go and backed by PostgreSQL. It allows users to register accounts, subscribe to RSS feeds, periodically scrape and store feed posts in the background, and browse posts directly from the terminal.

---

## Prerequisites

Before running Gator, ensure you have the following installed on your system:

1. **Go** (version 1.22 or higher)
   - Verify installation: `go version`
2. **PostgreSQL** (running locally or accessible via network)
   - Verify installation: `psql --version`
   - Ensure the PostgreSQL service is active and you have permissions to create a database.
3. *(Optional)* **Goose** database migration tool:
   - Install goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`

---

## Database Setup

1. **Create the Database:**
   ```bash
   createdb gator
   ```
   *(Or with PostgreSQL prompt: `CREATE DATABASE gator;`)*

2. **Run Migrations:**
   Apply database schema migrations from the `sql/schema` directory using `goose`:
   ```bash
   cd sql/schema
   goose postgres "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable" up
   cd ../..
   ```

---

## Installation

Install the `gator` CLI binary using `go install`:

### From the cloned repository root:
```bash
go install .
```

### Or directly via Go package path:
```bash
go install github.com/dontsitdowncauseimovedyourchair/gator@latest
```

---

## Configuration

Gator requires a JSON configuration file named `.gatorconfig.json` placed in your home directory (`~/.gatorconfig.json`).

Create `~/.gatorconfig.json` with the following structure:

```json
{
  "db_url": "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

- `db_url`: The PostgreSQL connection string used by Gator.
- `current_user_name`: The active logged-in user (automatically updated when running `login` or `register`).

---

## Command Reference

Run `gator <command> [arguments]` from the terminal.

### User Management

| Command | Usage | Description |
| :--- | :--- | :--- |
| **`register`** | `gator register <username>` | Creates a new user account in the database and logs in as that user. |
| **`login`** | `gator login <username>` | Sets the active user in `~/.gatorconfig.json` (user must already exist). |
| **`users`** | `gator users` | Lists all registered users and highlights the currently active user. |
| **`reset`** | `gator reset` | **Caution:** Clears all user records and associated feeds/posts from the database. |

### Feed Management

| Command | Usage | Description |
| :--- | :--- | :--- |
| **`addfeed`** | `gator addfeed <name> <url>` | Adds a new RSS feed and automatically follows it for the logged-in user. *(Requires login)* |
| **`feeds`** | `gator feeds` | Lists all feeds registered in the system along with creator information. |
| **`follow`** | `gator follow <url>` | Subscribes the logged-in user to an existing feed URL. *(Requires login)* |
| **`following`** | `gator following` | Lists all feeds currently followed by the logged-in user. *(Requires login)* |
| **`unfollow`** | `gator unfollow <url>` | Unsubscribes the logged-in user from a feed. *(Requires login)* |

### Aggregation & Reading

| Command | Usage | Description |
| :--- | :--- | :--- |
| **`agg`** | `gator agg <interval>` | Starts the long-running feed aggregator worker that continuously fetches feeds at the specified interval (e.g. `10s`, `1m`, `1h`). Minimum interval: `100ms`. |
| **`browse`** | `gator browse [limit]` | Displays the most recent posts from feeds followed by the logged-in user. `[limit]` defaults to 2 if omitted. *(Requires login)* |

---

## Quickstart Walkthrough

1. **Register a user:**
   ```bash
   gator register alice
   ```

2. **Add a feed:**
   ```bash
   gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"
   ```

3. **Start the aggregator (in a separate terminal or background):**
   ```bash
   gator agg 1m
   ```

4. **Browse latest articles:**
   ```bash
   gator browse 5
   ```
