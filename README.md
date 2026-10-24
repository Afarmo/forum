# forum
Literary Lions Forum is a full‑stack web forum built in Go with SQLite, designed to help book club members discuss literature, share insights, and organize conversations through categories.

## Running the Forum

### Run locally without Docker

From the project root, run:

```bash
make run
```

This runs the Go application locally with the SQLite FTS5 build tag. Make sure the required dependencies are installed.

### Running the Forum with Docker

#### Start the application

```bash
make up
```


### Access the Forum

Open [http://localhost:8000](http://localhost:8000) in your browser.


#### Rebuild and start the application

After changing Go code or the Dockerfile, run:

```bash
make build
```

#### Stop the application

```bash
make down
```

#### View logs

```bash
make logs
```

Press `Ctrl+C` to stop viewing the logs without stopping the application.

#### Check container status

```bash
make ps
```

#### Restart the application with a rebuild

```bash
make re
```

