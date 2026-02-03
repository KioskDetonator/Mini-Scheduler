# Mini-Scheduler

> A custom container orchestration engine built in Go. Utilizes the Docker SDK and Linux cgroups to schedule concurrent batch jobs, enforce resource limits, and monitor real-time node usage.

![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![Docker SDK](https://img.shields.io/badge/Docker%20SDK-v25-blue)
![License](https://img.shields.io/badge/License-MIT-green)

## 📖 Overview

Mini-Scheduler is a lightweight simulation of a cluster manager (like Kubernetes or Nomad). It was built to solve the challenge of executing high-throughput batch jobs on resource-constrained nodes.

Instead of relying on external orchestrators, this project implements the core scheduling logic from scratch using **Go's concurrency primitives** (`channels`, `goroutines`, `waitgroups`) and interfaces directly with the **Docker Engine API**.

## 🏗 Architecture

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