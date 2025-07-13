# Docker + Go + PostgreSQL Example

This project demonstrates how to run a simple Go web application connected to a PostgreSQL database using Docker Compose.

## Project Structure

```
.
├── docker-compose.yml
├── Dockerfile
├── src
│   └── main.go
```

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

## Usage

1. **Clone the repository** (if needed) and navigate to the project directory.

2. **Build and start the services:**
   ```sh
   docker-compose up --build
   ```

3. **Access the Go web app:**
   - Open your browser and go to [http://localhost:8080](http://localhost:8080)
   - You should see: `Hello, World! Connected to PostgreSQL database.`

## Configuration

- The Go app connects to the PostgreSQL database using environment variables defined in `docker-compose.yml`.
- The database data is persisted in a Docker volume (`postgres_data`).

## Stopping the Services

To stop and remove the containers, run:
```sh
docker-compose down
```

## License

MIT