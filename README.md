# forum
Literary Lions Forum is a full‑stack web forum built in Go with SQLite, designed to help book club members discuss literature, share insights, and organize conversations through categories.

## Running the Forum with Docker

### Start the application

From the project root, run:

```bash
docker compose up --build -d
```

Open the forum at http://localhost:8080.

### Stop the application

```bash
docker compose down
```

### View logs

```bash
docker compose logs -f
```

Press `Ctrl+C` to stop viewing the logs without stopping the application.

### Check container status

```bash
docker compose ps
```