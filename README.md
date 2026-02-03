# Mini-Scheduler

> A custom container orchestration engine built in Go. Utilizes the Docker SDK and Linux cgroups to schedule concurrent batch jobs, enforce resource limits, and monitor real-time node usage.

![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![Docker SDK](https://img.shields.io/badge/Docker%20SDK-v25-blue)
![License](https://img.shields.io/badge/License-MIT-green)

## Overview

Mini-Scheduler is a lightweight simulation of a cluster manager (like Kubernetes or Nomad). It was built to solve the challenge of executing high-throughput batch jobs on resource-constrained nodes.

Instead of relying on external orchestrators, this project implements the core scheduling logic from scratch using **Go's concurrency primitives** (`channels`, `goroutines`, `waitgroups`) and interfaces directly with the **Docker Engine API**.

## Architecture

The system uses a **Worker Pool** pattern to manage concurrency. A central Manager distributes tasks to a fixed number of workers, ensuring that the node is never overloaded, regardless of how many jobs are submitted.

```mermaid
graph TD
    User[User / CLI] -->|Submit 50+ Jobs| Queue(Job Queue)
    Queue -->|Pull Task| Scheduler{Scheduler Manager}
    
    subgraph "Worker Pool (Concurrency Limit: 3)"
        W1[Worker 1]
        W2[Worker 2]
        W3[Worker 3]
    end
    
    Scheduler -->|Dispatch| W1
    Scheduler -->|Dispatch| W2
    Scheduler -->|Dispatch| W3
    
    W1 -->|API Call| Docker[Docker Engine]
    W2 -->|API Call| Docker
    W3 -->|API Call| Docker
    
    Docker -.->|Stats (Cgroups)| Monitor[Resource Monitor]

```

## Key Features

* **Concurrency Management:** Implements a Round-Robin scheduling algorithm to process jobs using a thread-safe worker pool.
* **Resource Isolation:** Enforces hard limits (RAM/CPU) on every container via `HostConfig` to prevent noisy neighbor issues.
* **Real-Time Monitoring:** Queries Linux cgroups via the Docker API to track live memory and CPU usage of running tasks.
* **Graceful Lifecycle:** Handles container creation, execution, monitoring, and cleanup (garbage collection) automatically.

## Usage

### Prerequisites

Before running this project, you must have the following installed:

* **Go (1.21+)**: [Download Here](https://go.dev/dl/)
* **Docker Desktop**: [Download Here](https://www.docker.com/products/docker-desktop/) (Ensure WSL 2 integration is enabled in settings)

### 1. Installation

Clone the repository and install the dependencies.

```bash
# Clone the repo
git clone [https://github.com/YOUR_USERNAME/mini-scheduler.git](https://github.com/YOUR_USERNAME/mini-scheduler.git)

# Enter the directory
cd mini-scheduler

# Install Go dependencies
go mod tidy

```

### 2. Running the Scheduler

You can run the scheduler directly using `go run`.

**Basic Test:**
Runs 5 simple jobs using the Alpine image to verify concurrency.

```bash
go run main.go

```

**Custom Load Test:**
Override the task count, image, or command to simulate real workloads.

```bash
# Example: Run 10 Python containers to simulate CPU load
go run main.go -count=10 -image="python:3.9-slim" -cmd="python -c 'print(2**10000)'"

```

## Technical Highlights

* **Concurrency:** Used Go Channels as a semaphore pattern to strictly limit the number of active goroutines, preventing scheduler exhaustion.
* **Systems Programming:** Interacted with low-level kernel features (Cgroups) through the Docker socket (`/var/run/docker.sock`) to gather telemetry.
* **Clean Architecture:** Separated the `Scheduler` (logic), `Monitor` (observability), and `Worker` (execution) into distinct, testable components.

