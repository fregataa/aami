#!/usr/bin/env python3
"""
AAMI API Client Library

A Python client for interacting with the AAMI Config Server REST API.

Example usage:
    from api_client import AAMIClient

    client = AAMIClient(base_url="http://localhost:8080/api/v1")

    # Create a group
    group = client.create_group(
        name="production",
        namespace="environment",
        description="Production environment"
    )

    # Register a target
    target = client.create_target(
        hostname="gpu-node-01.example.com",
        ip_address="10.0.1.10",
        primary_group_id=group["id"],
        exporters=[
            {"type": "node_exporter", "port": 9100, "enabled": True}
        ]
    )
"""

import os
import json
from typing import Dict, List, Optional, Any
import requests


class AAMIClient:
    """Client for AAMI Config Server API"""

    def __init__(
        self,
        base_url: Optional[str] = None,
        api_key: Optional[str] = None,
        timeout: int = 30
    ):
        """
        Initialize AAMI API client

        Args:
            base_url: API base URL (default: from AAMI_API_URL env var)
            api_key: API key for authentication (default: from AAMI_API_KEY env var)
            timeout: Request timeout in seconds (default: 30)
        """
        self.base_url = base_url or os.getenv("AAMI_API_URL", "http://localhost:8080/api/v1")
        self.api_key = api_key or os.getenv("AAMI_API_KEY")
        self.timeout = timeout
        self.session = requests.Session()

        # Set default headers
        self.session.headers.update({"Content-Type": "application/json"})
        if self.api_key:
            self.session.headers.update({"Authorization": f"Bearer {self.api_key}"})

    def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict] = None,
        params: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """
        Make HTTP request to API

        Args:
            method: HTTP method (GET, POST, PUT, DELETE)
            endpoint: API endpoint path
            data: Request body data
            params: Query parameters

        Returns:
            Response JSON data

        Raises:
            requests.HTTPError: If request fails
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"

        response = self.session.request(
            method=method,
            url=url,
            json=data,
            params=params,
            timeout=self.timeout
        )

        response.raise_for_status()
        return response.json()

    # Health Check

    def health(self) -> Dict[str, Any]:
        """Get API health status"""
        return self._request("GET", "/health")

    # Groups API

    def list_groups(
        self,
        namespace: Optional[str] = None,
        parent_id: Optional[str] = None,
        page: int = 1,
        limit: int = 50
    ) -> Dict[str, Any]:
        """
        List all groups

        Args:
            namespace: Filter by namespace
            parent_id: Filter by parent group
            page: Page number
            limit: Items per page
        """
        params = {"page": page, "limit": limit}
        if namespace:
            params["namespace"] = namespace
        if parent_id:
            params["parent_id"] = parent_id

        return self._request("GET", "/groups", params=params)

    def get_group(self, group_id: str) -> Dict[str, Any]:
        """Get group by ID"""
        return self._request("GET", f"/groups/{group_id}")

    def create_group(
        self,
        name: str,
        namespace: str,
        description: str,
        parent_id: Optional[str] = None,
        metadata: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """
        Create a new group

        Args:
            name: Group name
            namespace: Namespace (infrastructure, logical, environment)
            description: Group description
            parent_id: Parent group ID (optional)
            metadata: Additional metadata (optional)
        """
        data = {
            "name": name,
            "namespace": namespace,
            "description": description,
            "parent_id": parent_id,
            "metadata": metadata or {}
        }
        return self._request("POST", "/groups", data=data)

    def update_group(
        self,
        group_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        metadata: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """Update group"""
        data = {}
        if name:
            data["name"] = name
        if description:
            data["description"] = description
        if metadata:
            data["metadata"] = metadata

        return self._request("PUT", f"/groups/{group_id}", data=data)

    def delete_group(self, group_id: str) -> Dict[str, Any]:
        """Delete group"""
        return self._request("DELETE", f"/groups/{group_id}")

    # Targets API

    def list_targets(
        self,
        group_id: Optional[str] = None,
        status: Optional[str] = None,
        page: int = 1,
        limit: int = 50
    ) -> Dict[str, Any]:
        """List all targets"""
        params = {"page": page, "limit": limit}
        if group_id:
            params["group_id"] = group_id
        if status:
            params["status"] = status

        return self._request("GET", "/targets", params=params)

    def get_target(self, target_id: str) -> Dict[str, Any]:
        """Get target by ID"""
        return self._request("GET", f"/targets/{target_id}")

    def create_target(
        self,
        hostname: str,
        ip_address: str,
        primary_group_id: str,
        exporters: List[Dict],
        labels: Optional[Dict] = None,
        metadata: Optional[Dict] = None,
        secondary_group_ids: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Register a new target

        Args:
            hostname: Target hostname
            ip_address: Target IP address
            primary_group_id: Primary group ID
            exporters: List of exporter configurations
            labels: Target labels (optional)
            metadata: Additional metadata (optional)
            secondary_group_ids: Additional group memberships (optional)
        """
        data = {
            "hostname": hostname,
            "ip_address": ip_address,
            "primary_group_id": primary_group_id,
            "exporters": exporters,
            "labels": labels or {},
            "metadata": metadata or {},
            "secondary_group_ids": secondary_group_ids or []
        }
        return self._request("POST", "/targets", data=data)

    def update_target(
        self,
        target_id: str,
        labels: Optional[Dict] = None,
        metadata: Optional[Dict] = None,
        exporters: Optional[List[Dict]] = None
    ) -> Dict[str, Any]:
        """Update target"""
        data = {}
        if labels:
            data["labels"] = labels
        if metadata:
            data["metadata"] = metadata
        if exporters:
            data["exporters"] = exporters

        return self._request("PUT", f"/targets/{target_id}", data=data)

    def delete_target(self, target_id: str) -> Dict[str, Any]:
        """Delete target"""
        return self._request("DELETE", f"/targets/{target_id}")

    # Alert Rules API

    def list_alert_templates(self) -> Dict[str, Any]:
        """List available alert rule templates"""
        return self._request("GET", "/alert-templates")

    def apply_alert_rule(
        self,
        group_id: str,
        rule_template_id: str,
        config: Dict,
        enabled: bool = True,
        merge_strategy: str = "override"
    ) -> Dict[str, Any]:
        """
        Apply alert rule to group

        Args:
            group_id: Group ID
            rule_template_id: Alert rule template ID
            config: Rule configuration (thresholds, duration, etc.)
            enabled: Whether rule is enabled
            merge_strategy: Merge strategy (override, merge)
        """
        data = {
            "rule_template_id": rule_template_id,
            "enabled": enabled,
            "config": config,
            "merge_strategy": merge_strategy
        }
        return self._request("POST", f"/groups/{group_id}/alert-rules", data=data)

    def get_effective_alert_rules(self, target_id: str) -> Dict[str, Any]:
        """Get effective alert rules for target"""
        return self._request("GET", f"/targets/{target_id}/alert-rules/effective")

    def trace_alert_policy(self, target_id: str) -> Dict[str, Any]:
        """Trace alert rule policy inheritance for target"""
        return self._request("GET", f"/targets/{target_id}/alert-rules/trace")

    # Service Discovery API

    def get_prometheus_sd(self) -> List[Dict]:
        """Get Prometheus service discovery targets"""
        return self._request("GET", "/sd/prometheus")

    def get_alert_rules(self) -> Dict[str, Any]:
        """Get Prometheus alert rules"""
        return self._request("GET", "/sd/alert-rules")

    # Bootstrap API

    def create_bootstrap_token(
        self,
        name: str,
        expires_at: str,
        max_uses: int,
        default_group_id: str,
        labels: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """Create bootstrap token for auto-registration"""
        data = {
            "name": name,
            "expires_at": expires_at,
            "max_uses": max_uses,
            "default_group_id": default_group_id,
            "labels": labels or {}
        }
        return self._request("POST", "/bootstrap/tokens", data=data)


# Example usage
if __name__ == "__main__":
    # Initialize client
    client = AAMIClient()

    # Check health
    print("Checking API health...")
    health = client.health()
    print(f"Status: {health['status']}")

    # Create a group
    print("\nCreating group...")
    group = client.create_group(
        name="example-group",
        namespace="environment",
        description="Example group for testing"
    )
    print(f"Created group: {group['id']}")

    # Register a target
    print("\nRegistering target...")
    target = client.create_target(
        hostname="example-node.local",
        ip_address="192.168.1.100",
        primary_group_id=group["id"],
        exporters=[
            {
                "type": "node_exporter",
                "port": 9100,
                "enabled": True
            }
        ],
        labels={
            "environment": "test"
        }
    )
    print(f"Registered target: {target['id']}")

    # List targets
    print("\nListing targets...")
    targets = client.list_targets()
    print(f"Total targets: {targets['total']}")
