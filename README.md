# telescope

Expose custom metrics from Weave Scope.

## How it works

Collect data from Weave Scope with Topology/Node APIs, normalize & expose data to Prometheus via GET /metrics endpoint.

## Definition

- **src**: host/server that sends a going-out connection.
- **src_ns**: src namespace
- **dest**: host/server that receives a coming-in connection.
- **dest_ns**: dest namespace
- **dest_port**: destination port
