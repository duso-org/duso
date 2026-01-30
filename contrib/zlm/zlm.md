# zlm - Zero Language Model

A test utility that simulates LLM output without burning tokens.

## Purpose

Generate plausible-looking gibberish text with configurable delays to test Duso's concurrency model and worker swarms at scale without making actual LLM API calls.

Perfect for testing:
- Large-scale worker spawning (tested up to 100,000 concurrent workers)
- Datastore coordination patterns
- Swarm orchestration logic
- Latency-sensitive synchronization

## Usage

```duso
zlm = require("zlm")

// Generate mock output with default params (200 tokens, 1.5s delay)
text = zlm.prompt()

// Custom parameters
text = zlm.prompt(tokens=500, delay=2.5)
```

## Parameters

- **tokens** (default: 200) - Approximate number of tokens to generate (varies by ±20%)
- **delay** (default: 1.5) - Simulated processing time in seconds (varies by ±20%)
- **id** - Not currently used (reserved for future use)

## How It Works

- Generates text using weighted character distribution that creates word-salad gibberish
- Adds natural variation to both token count and delay time
- Uses `sleep()` to simulate real processing latency
- Characters are weighted to produce roughly natural-looking output (common letters more frequent)

## Example: Worker Swarm

```duso
zlm = require("zlm")
store = datastore("workers")

ctx = context()

if ctx == nil then
  // Main: spawn workers
  store.set("completed", 0)

  for i = 1, 1000 do
    spawn("swarm_worker.du")
  end

  // Wait for all to finish
  store.wait("completed", 1000)
  print("All workers done!")
end

// Worker code runs when ctx != nil
text = zlm.prompt(tokens=300, delay=1.0)
store.increment("completed", 1)
exit({text = text})
```

## Notes

- Generates different gibberish each run (uses random)
- Character distribution roughly mimics English (more common letters weighted higher)
- Useful for load testing without API rate limits or costs
