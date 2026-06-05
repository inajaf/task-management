# Task

Develop a task management mechanism accessible via HTTP.
The system provides the following capabilities:
- Users can create new tasks via an HTTP endpoint.
- Newly created tasks are added to a queue for asynchronous processing.
- Users can list all tasks and check the status of individual tasks using their task ID.
- The system includes both a task queue and a task management component to handle task execution and status tracking.

## Run the service:
```bash
make run
```

## Developer Commands:
```bash
make fmt   # Format the Go source code
make vet   # Run static analysis (go vet linter)
make test  # Run all unit and handler tests
```

## Usage (Requests)

### 1. Create Tasks (Mock or Custom)
* **Endpoint**: `POST - http://localhost:2025/task` (also supports legacy path `/task/create`)
* **Mock Data Fallback**: Send an empty request body to generate mock tasks automatically.
* **Custom Task Input**: Send a JSON body to create a specific task:
  ```json
  {
    "user_id": 42,
    "user_name": "Alice",
    "task_id": 999,
    "name": "My Custom Task"
  }
  ```

### 2. List Tasks
* **Endpoint**: `GET - http://localhost:2025/task/create` (also supports standard `/tasks` path)
* **Response**: Returns a list of all tasks and their current statuses.

### 3. Get Task Status by ID
* **Endpoint**: `GET - http://localhost:2025/tasks/3`
* **Response**: Returns the status of the specific task (e.g., `Created`, `Processing`, `Done`).

### 4. Update Task Status
* **Endpoint**: `PUT - http://localhost:2025/tasks/3`
* **Response**: Updates the status of the task manually to `Processing` (while background workers transition them to `Done`).

---
*Note: Task storage is fully implemented using thread-safe maps (`sync.RWMutex`) to guarantee safe concurrent execution.*
