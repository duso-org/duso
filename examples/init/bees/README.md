# Bees Template

A swarm coordination example showcasing Duso's concurrency and state management:

- **spawn()** - Launch background worker processes
- **datastore()** - Thread-safe, in-memory key-value store
- **increment()** - Atomic operations for safe counting
- **context()** - Access spawn context and request data
- **Distributed coordination** - Workers without shared memory

## Running

```bash
duso bees.du
```

The orchestrator spawns 10 worker bees, each doing 5 units of work and updating shared counters via the datastore. After 2 seconds, it reports results.

## How It Works

1. **main.du** - Initializes the swarm datastore and spawns workers
2. **worker.du** - Each bee increments shared counters atomically
3. Counters are updated safely without locks or shared memory

## Learn More

- [spawn() documentation](/docs/reference/spawn.md)
- [datastore() documentation](/docs/reference/datastore.md)
- [Swarm coordination patterns](/docs/learning-duso.md#coordinating-worker-swarms)
