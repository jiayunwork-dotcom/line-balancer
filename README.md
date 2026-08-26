# line-balancer

Production line balancing and scheduling toolkit. Computes takt time, assigns
tasks to stations using multiple heuristics (greedy best-fit, Ranked Positional
Weight, Kilbridge-Wester, Moodie-Young, dual-sided), and evaluates line
efficiency, smoothness, and throughput metrics.

## Capabilities

- Task precedence graph modeling with DAG validation and critical-path analysis
- Multiple balance algorithms respecting precedence constraints
- Zone and incompatibility constraints
- Mixed-model line sequencing with changeover optimization
- Multi-objective comparison of balance solutions (efficiency, smoothness, station count)

## Usage

```bash
# Simple balance (no precedence):
go run . -in example/tasks.csv -demand 400 -time 28800

# With precedence constraints:
go run . -prec example/precedence.csv -demand 400 -time 28800 -method rpw

# Available methods: greedy, rpw, kw, moodie
```

## Input format

Simple CSV (`-in`):
```
task,seconds
weld,45
paint,30
```

Precedence CSV (`-prec`):
```
task_id,seconds,predecessors
cut,40,-
weld,45,cut
assemble,60,weld;polish
```

## Build and test

```bash
go build ./...
go test ./...
```

## Docker

```bash
docker build -t line-balancer .
docker run --rm line-balancer go test ./...
docker run --rm -v "$PWD":/src line-balancer go run . -in /src/example/tasks.csv -demand 400
```

## License

MIT
