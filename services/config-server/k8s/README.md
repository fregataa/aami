# Kubernetes Manifests for AAMI Config Server

This directory contains Kubernetes manifests for deploying the AAMI Config Server to a Kubernetes cluster.

## Files

- **namespace.yaml**: Creates the `aami` namespace
- **configmap.yaml**: Configuration settings (non-sensitive)
- **secret.yaml**: Sensitive configuration (credentials)
- **deployment.yaml**: Config Server Deployment with 3 replicas
- **service.yaml**: ClusterIP Service exposing port 80
- **ingress.yaml**: Ingress for external access
- **hpa.yaml**: HorizontalPodAutoscaler for auto-scaling
- **kustomization.yaml**: Kustomize configuration

## Prerequisites

- Kubernetes cluster (v1.25+)
- kubectl configured
- Ingress controller installed (e.g., NGINX Ingress Controller)
- PostgreSQL database (external or in-cluster)

## Deployment

### Using kubectl

```bash
# Apply all manifests
kubectl apply -f k8s/

# Or apply individually in order
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml
```

### Using Kustomize

```bash
# Apply with kustomize
kubectl apply -k k8s/

# Preview changes
kubectl kustomize k8s/
```

## Configuration

### Update Secrets

Before deploying to production, update the secrets:

```bash
# Edit secret.yaml and change passwords
vi k8s/secret.yaml

# Or create secret from command line
kubectl create secret generic config-server-secret \
  --from-literal=DB_USER=postgres \
  --from-literal=DB_PASSWORD=your-secure-password \
  -n aami
```

### Update ConfigMap

Modify `configmap.yaml` to match your environment:

- Database host/port
- Log level

### Update Ingress

Modify `ingress.yaml`:

- Change host to your domain
- Configure TLS if needed
- Adjust rate limiting

## Verification

```bash
# Check deployment status
kubectl get deployments -n aami
kubectl get pods -n aami
kubectl get services -n aami
kubectl get ingress -n aami

# Check logs
kubectl logs -f deployment/config-server -n aami

# Check health
kubectl port-forward svc/config-server 8080:80 -n aami
curl http://localhost:8080/health
```

## Scaling

### Manual Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment config-server --replicas=5 -n aami
```

### Auto-scaling (HPA)

The HorizontalPodAutoscaler is configured to:
- Minimum replicas: 3
- Maximum replicas: 10
- Target CPU: 70%
- Target Memory: 80%

```bash
# Check HPA status
kubectl get hpa -n aami
kubectl describe hpa config-server -n aami
```

## Troubleshooting

### Pod not starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n aami

# Check logs
kubectl logs <pod-name> -n aami
```

### Database connection issues

```bash
# Check if DB is reachable
kubectl exec -it <pod-name> -n aami -- /bin/sh
# Inside pod:
wget -O- http://postgres.aami.svc.cluster.local:5432 2>&1 | head
```

### Health check failures

```bash
# Test health endpoints
kubectl port-forward svc/config-server 8080:80 -n aami
curl http://localhost:8080/health
curl http://localhost:8080/health/ready
curl http://localhost:8080/health/live
```

## Production Recommendations

1. **External Database**: Use managed PostgreSQL (AWS RDS, GCP Cloud SQL, etc.)
2. **Secret Management**: Use external secret management (HashiCorp Vault, AWS Secrets Manager)
3. **TLS**: Enable TLS in Ingress with cert-manager
4. **Monitoring**: Install Prometheus and Grafana for monitoring
5. **Logging**: Configure centralized logging (ELK, Loki)
6. **Backup**: Set up regular database backups
7. **Resource Limits**: Adjust resource requests/limits based on load testing
8. **Network Policies**: Implement NetworkPolicies for network segmentation

## Cleanup

```bash
# Delete all resources
kubectl delete -f k8s/

# Or delete namespace (removes everything)
kubectl delete namespace aami
```
