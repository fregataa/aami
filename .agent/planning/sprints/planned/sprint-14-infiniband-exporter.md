# Sprint 13: InfiniBand Network Exporter

**Status**: 📋 Planned
**Duration**: 2-3 weeks
**Started**: TBD
**Completed**: TBD

## Goals
Implement InfiniBand high-speed network monitoring exporter.

## Background
InfiniBand is a high-speed, low-latency interconnect used in AI/HPC clusters for inter-node communication and RDMA operations.

## Metrics to Collect

### Port Statistics
- Data received/transmitted (bytes)
- Packets received/transmitted
- Multicast packets
- Link state (active/down)
- Physical state
- Link width and speed

### Performance Metrics
- Port utilization (%)
- Packet error rate
- Symbol errors
- Link recovery count
- VL15 dropped packets

### RDMA Metrics
- RDMA read/write operations
- RDMA atomic operations
- RDMA latency

## Tasks

### Research & Design (2 days)
- [ ] Research InfiniBand metrics collection
- [ ] Test `perfquery` command
- [ ] Analyze `/sys/class/infiniband/` structure
- [ ] Study OFED tools
- [ ] Design metric schema

### Implementation (5 days)
- [ ] Implement perfquery wrapper
- [ ] Implement sysfs parser
- [ ] Create metric collectors
- [ ] Support multiple HCAs
- [ ] Support multiple ports
- [ ] Implement topology discovery
- [ ] Add fabric-wide metrics

### Testing (2 days)
- [ ] Write unit tests
- [ ] Mock InfiniBand devices
- [ ] Write integration tests
- [ ] Test with real IB fabric

### Deployment (2 days)
- [ ] Create Dockerfile
- [ ] Create Kubernetes DaemonSet (privileged)
- [ ] Create Ansible role
- [ ] Write deployment guide

### Monitoring (2 days)
- [ ] Create Grafana dashboard
- [ ] Create alert rules
- [ ] Test monitoring setup

## Deliverables
- `infiniband-exporter` binary
- Docker image
- Kubernetes DaemonSet
- Grafana dashboard
- Alert rules
- Documentation

## Example Metrics
```
infiniband_port_receive_data_bytes_total{port="1",device="mlx5_0"} 1234567890
infiniband_port_transmit_data_bytes_total{port="1",device="mlx5_0"} 987654321
infiniband_port_state{port="1",device="mlx5_0",state="active"} 1
infiniband_port_rate_gbps{port="1",device="mlx5_0"} 100
infiniband_port_errors_total{port="1",device="mlx5_0",type="symbol"} 0
```

## Success Criteria
- Exporter collects all port metrics
- Dashboard shows network health
- Alerts detect link issues
- Supports multiple IB vendors (Mellanox, Intel)

## Security Considerations
- Requires privileged access or CAP_SYS_ADMIN
- Run in separate security context in K8s

## Notes
- Test with different IB hardware
- Handle offline ports gracefully
- Monitor collection performance
