\# Gocache: Distributed Cache with Owner-Based Routing



A production-style distributed in-memory cache built in Go, featuring consistent hashing, owner-based load semantics, and gRPC-based inter-node communication.



\---



\## 🚀 Overview



Gocache is a distributed cache system where:



\- Requests can enter any node

\- Keys are routed to owner nodes via consistent hashing

\- Only the owner node performs load-on-miss

\- Nodes communicate via gRPC for efficient internal RPC



\---



\## 🧠 Architecture



Client → Node A (HTTP entry) → gRPC → Node B (owner)



\- HTTP server: handles client requests

\- gRPC server: handles inter-node communication

\- Owner node: performs backend load and stores cache



\---



\## ✨ Key Features



\### 🔹 Owner-Based Load-on-Miss

Only the responsible node loads data from backend, preventing duplicate loads across nodes.



\### 🔹 Consistent Hashing

Distributes keys across nodes with minimal remapping during scaling.



\### 🔹 gRPC Communication

Efficient binary RPC for node-to-node interaction.



\### 🔹 LRU + TTL Cache

\- LRU eviction

\- TTL expiration

\- Memory-safe design



\### 🔹 Singleflight

Prevents cache stampede under high concurrency.



\---



\## ⚙️ Tech Stack



\- Go (goroutines, sync primitives)

\- gRPC + Protobuf

\- Consistent Hashing

\- LRU Cache

\- HTTP + gRPC dual-server architecture



\---



\## 🧪 Demo



\### 1️⃣ Start Node B (owner first)





go run ./cmd/server -addr=:8082 -grpc\_addr=:9092 -self=nodeB -transport=grpc -peers="nodeA=localhost:9091,nodeB=localhost:9092"





\### 2️⃣ Start Node A





go run ./cmd/server -addr=:8081 -grpc\_addr=:9091 -self=nodeA -transport=grpc -peers="nodeA=localhost:9091,nodeB=localhost:9092"





\---



\### 3️⃣ Send request





curl "http://localhost:8081/get?key=user:123

"





\---



\### 4️⃣ Example Response





{"key":"user:123","node":"nodeA","value":"db-value-for-user:123-from-nodeB"}





\---



\### 🔥 Key Observation



Even though the request hits \*\*nodeA\*\*, the actual load happens on \*\*nodeB\*\*.



Check logs:





OWNER LOAD for key: user:123 on nodeB





\---



\## 📈 What This Demonstrates



\- Distributed key routing

\- Owner-based execution model

\- Separation of client entry and system communication

\- Real multi-node coordination via gRPC



\---



\## 🔮 Future Work



\- Service discovery (etcd)

\- Prometheus metrics

\- Circuit breaker / retry logic

\- Kubernetes deployment

\- Replication \& fault tolerance



\---



\## 🧑‍💻 Author



Built as a distributed systems project focusing on backend infrastructure and scalable architecture.

