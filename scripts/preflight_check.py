#!/usr/bin/env python3
"""
AAMI Preflight Check Script

Validates system requirements before AAMI installation to prevent
mid-install failures. Supports both server and node installation modes.

Usage:
    ./preflight_check.py [OPTIONS]

Options:
    --mode MODE      Check mode: 'server' or 'node' (default: auto-detect)
    --server URL     Config Server URL (for node mode connectivity check)
    --fix            Attempt automatic fixes for issues
    --json           Output results in JSON format
    --quiet          Only show errors
    --verbose        Show detailed check information
    -h, --help       Show this help message
"""

import argparse
import json
import os
import platform
import shutil
import socket
import subprocess
import sys
from dataclasses import dataclass, field
from datetime import datetime, timezone
from enum import Enum
from pathlib import Path
from typing import Optional

VERSION = "1.0.0"


class CheckStatus(Enum):
    PASS = "pass"
    FAIL = "fail"
    WARN = "warn"
    INFO = "info"


class Mode(Enum):
    SERVER = "server"
    NODE = "node"


@dataclass
class Requirements:
    """System requirements for each mode."""
    min_cpu: int
    min_ram_gb: int
    min_disk_gb: int
    ports: list[int]


SERVER_REQUIREMENTS = Requirements(
    min_cpu=2,
    min_ram_gb=4,
    min_disk_gb=20,
    ports=[8080, 9090, 3000, 5432, 6379],
)

NODE_REQUIREMENTS = Requirements(
    min_cpu=1,
    min_ram_gb=1,
    min_disk_gb=5,
    ports=[9100, 9400],
)


@dataclass
class CheckResult:
    """Result of a single check."""
    name: str
    status: CheckStatus
    message: str
    details: dict = field(default_factory=dict)


@dataclass
class PreflightResults:
    """Aggregated results from all checks."""
    version: str = VERSION
    mode: str = ""
    timestamp: str = ""
    passed: bool = True
    checks: dict = field(default_factory=dict)
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    fixes: list[str] = field(default_factory=list)


class Colors:
    """ANSI color codes for terminal output."""
    RED = "\033[0;31m"
    GREEN = "\033[0;32m"
    YELLOW = "\033[1;33m"
    BLUE = "\033[0;34m"
    CYAN = "\033[0;36m"
    BOLD = "\033[1m"
    NC = "\033[0m"  # No Color

    @classmethod
    def disable(cls) -> None:
        """Disable colors for non-terminal or JSON output."""
        cls.RED = ""
        cls.GREEN = ""
        cls.YELLOW = ""
        cls.BLUE = ""
        cls.CYAN = ""
        cls.BOLD = ""
        cls.NC = ""


class PreflightChecker:
    """Main preflight check orchestrator."""

    def __init__(
        self,
        mode: Optional[str] = None,
        config_server_url: str = "",
        fix_mode: bool = False,
        json_output: bool = False,
        quiet_mode: bool = False,
        verbose_mode: bool = False,
    ) -> None:
        self.mode = Mode(mode) if mode else None
        self.config_server_url = config_server_url
        self.fix_mode = fix_mode
        self.json_output = json_output
        self.quiet_mode = quiet_mode
        self.verbose_mode = verbose_mode
        self.results = PreflightResults()

        if json_output or not sys.stdout.isatty():
            Colors.disable()

    def run(self) -> int:
        """Run all preflight checks and return exit code."""
        self._detect_mode()
        self.results.mode = self.mode.value
        self.results.timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

        self._print_header()

        requirements = SERVER_REQUIREMENTS if self.mode == Mode.SERVER else NODE_REQUIREMENTS

        # System Requirements
        self._print_section("System Requirements")
        self._check_os()
        self._check_cpu(requirements.min_cpu)
        self._check_ram(requirements.min_ram_gb)
        self._check_disk(requirements.min_disk_gb)

        # Software Dependencies
        if self.mode == Mode.SERVER:
            self._check_software_server()
        else:
            self._check_software_node()

        # Network Connectivity
        if self.mode == Mode.SERVER:
            self._check_network_server()
        else:
            self._check_network_node()

        # Port Availability
        self._print_section("Port Availability")
        for port in requirements.ports:
            self._check_port(port)

        # Permissions
        if self.mode == Mode.SERVER:
            self._check_permissions_server()
        else:
            self._check_permissions_node()

        # Hardware Detection (node mode only)
        if self.mode == Mode.NODE:
            self._check_hardware()

        # Output results
        self.results.passed = len(self.results.errors) == 0

        if self.json_output:
            self._output_json()
        else:
            self._print_summary()

        return 0 if self.results.passed else 1

    def _detect_mode(self) -> None:
        """Auto-detect installation mode if not specified."""
        if self.mode:
            return

        # Check if Docker is available (likely server)
        if shutil.which("docker"):
            # Check for docker compose
            try:
                subprocess.run(
                    ["docker", "compose", "version"],
                    capture_output=True,
                    check=True,
                )
                self.mode = Mode.SERVER
                self._verbose("Auto-detected mode: server (Docker and Docker Compose available)")
                return
            except (subprocess.CalledProcessError, FileNotFoundError):
                pass

            if shutil.which("docker-compose"):
                self.mode = Mode.SERVER
                self._verbose("Auto-detected mode: server (Docker Compose standalone available)")
                return

        self.mode = Mode.NODE
        self._verbose("Auto-detected mode: node")

    # =========================================================================
    # System Requirements Checks
    # =========================================================================

    def _check_os(self) -> None:
        """Check operating system compatibility."""
        self._verbose("Checking operating system...")

        system = platform.system()
        os_info = {"name": "", "version": "", "pretty": "", "supported": False}

        if system == "Darwin":
            os_info["name"] = "macos"
            os_info["version"] = platform.mac_ver()[0]
            os_info["pretty"] = f"macOS {os_info['version']}"
            os_info["supported"] = False

            self._print_check(
                CheckStatus.WARN,
                f"OS: {os_info['pretty']} (development only, not for production)",
            )
            self.results.warnings.append(
                "macOS is supported for development/testing only. Production deployment requires Linux."
            )
            self.results.checks["os"] = os_info
            return

        if system != "Linux":
            self._print_check(CheckStatus.FAIL, f"OS: {system} (not supported)")
            self.results.errors.append(f"Operating system {system} is not supported")
            os_info["supported"] = False
            self.results.checks["os"] = os_info
            return

        # Parse /etc/os-release
        os_release = self._parse_os_release()
        os_info["name"] = os_release.get("ID", "unknown")
        os_info["version"] = os_release.get("VERSION_ID", "0")
        os_info["pretty"] = os_release.get("PRETTY_NAME", f"{os_info['name']} {os_info['version']}")

        # Check supported versions
        supported = True
        name = os_info["name"].lower()
        try:
            version_major = int(os_info["version"].split(".")[0])
        except (ValueError, IndexError):
            version_major = 0

        if name == "ubuntu" and version_major < 20:
            supported = False
        elif name == "debian" and version_major < 11:
            supported = False
        elif name in ("centos", "rocky", "rhel", "almalinux") and version_major < 8:
            supported = False
        elif name not in ("ubuntu", "debian", "centos", "rocky", "rhel", "almalinux"):
            self._print_check(CheckStatus.WARN, f"OS: {os_info['pretty']} (not officially tested)")
            self.results.warnings.append(f"OS {name} is not officially tested")
            os_info["supported"] = True
            self.results.checks["os"] = os_info
            return

        os_info["supported"] = supported
        self.results.checks["os"] = os_info

        if supported:
            self._print_check(CheckStatus.PASS, f"OS: {os_info['pretty']} (supported)")
        else:
            self._print_check(CheckStatus.FAIL, f"OS: {os_info['pretty']} (not supported)")
            self.results.errors.append(
                f"OS version {name} {os_info['version']} is not supported. "
                "Minimum: Ubuntu 20.04, Debian 11, CentOS/Rocky 8"
            )

    def _check_cpu(self, min_cpu: int) -> None:
        """Check CPU core count."""
        self._verbose("Checking CPU cores...")

        cpu_cores = os.cpu_count() or 0
        passed = cpu_cores >= min_cpu

        self.results.checks["cpu"] = {
            "cores": cpu_cores,
            "minimum": min_cpu,
            "passed": passed,
        }

        if passed:
            self._print_check(CheckStatus.PASS, f"CPU: {cpu_cores} cores (minimum: {min_cpu})")
        else:
            self._print_check(CheckStatus.FAIL, f"CPU: {cpu_cores} cores (minimum: {min_cpu})")
            self.results.errors.append(f"Insufficient CPU cores: {cpu_cores} (minimum: {min_cpu})")

    def _check_ram(self, min_ram_gb: int) -> None:
        """Check available RAM."""
        self._verbose("Checking RAM...")

        ram_gb = self._get_total_ram_gb()
        passed = ram_gb >= min_ram_gb

        self.results.checks["ram"] = {
            "gb": ram_gb,
            "minimum": min_ram_gb,
            "passed": passed,
        }

        if passed:
            self._print_check(CheckStatus.PASS, f"RAM: {ram_gb}GB (minimum: {min_ram_gb}GB)")
        else:
            self._print_check(CheckStatus.FAIL, f"RAM: {ram_gb}GB (minimum: {min_ram_gb}GB)")
            self.results.errors.append(f"Insufficient RAM: {ram_gb}GB (minimum: {min_ram_gb}GB)")

    def _check_disk(self, min_disk_gb: int) -> None:
        """Check available disk space."""
        self._verbose("Checking disk space...")

        disk_gb = self._get_available_disk_gb()
        passed = disk_gb >= min_disk_gb

        self.results.checks["disk"] = {
            "available_gb": disk_gb,
            "minimum": min_disk_gb,
            "passed": passed,
        }

        if passed:
            self._print_check(CheckStatus.PASS, f"Disk: {disk_gb}GB free (minimum: {min_disk_gb}GB)")
        else:
            self._print_check(CheckStatus.FAIL, f"Disk: {disk_gb}GB free (minimum: {min_disk_gb}GB)")
            self.results.errors.append(f"Insufficient disk space: {disk_gb}GB (minimum: {min_disk_gb}GB)")

    # =========================================================================
    # Software Dependency Checks
    # =========================================================================

    def _check_software_server(self) -> None:
        """Check software dependencies for server mode."""
        self._print_section("Software Dependencies")

        self._check_docker()
        self._check_docker_compose()
        self._check_command("curl", required=True)
        self._check_command("jq", required=False)

    def _check_software_node(self) -> None:
        """Check software dependencies for node mode."""
        self._print_section("Software Dependencies")

        self._check_command("curl", required=True)
        self._check_command("systemctl", required=True)
        self._check_command("tar", required=True)
        self._check_command("jq", required=False)

    def _check_command(self, cmd: str, required: bool = True) -> bool:
        """Check if a command is available."""
        path = shutil.which(cmd)
        sw_key = f"sw_{cmd}"

        if path:
            version = self._get_command_version(cmd)
            self.results.checks[sw_key] = {"installed": True, "version": version}
            self._print_check(CheckStatus.PASS, f"{cmd}: {version}")
            return True
        else:
            self.results.checks[sw_key] = {"installed": False}
            if required:
                self._print_check(CheckStatus.FAIL, f"{cmd}: not installed (required)")
                self.results.errors.append(f"{cmd} is required but not installed")
                self.results.fixes.append(f"Install {cmd}")
            else:
                self._print_check(CheckStatus.WARN, f"{cmd}: not installed (optional)")
                self.results.warnings.append(f"{cmd} is not installed (optional)")
            return False

    def _check_docker(self) -> bool:
        """Check Docker installation and daemon status."""
        self._verbose("Checking Docker...")

        if not self._check_command("docker", required=True):
            if self.fix_mode:
                self._info("Attempting to install Docker...")
                self._try_install_docker()
            return False

        # Check if Docker daemon is running
        try:
            subprocess.run(
                ["docker", "info"],
                capture_output=True,
                check=True,
                timeout=10,
            )
            self._print_check(CheckStatus.PASS, "Docker daemon: running")
            return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
            self._print_check(CheckStatus.FAIL, "Docker daemon: not running")
            self.results.errors.append("Docker daemon is not running")

            if self.fix_mode:
                self._info("Attempting to start Docker...")
                self._try_start_docker()
            else:
                self.results.fixes.append("Start Docker: sudo systemctl start docker")
            return False

    def _check_docker_compose(self) -> bool:
        """Check Docker Compose installation."""
        self._verbose("Checking Docker Compose...")

        # Check for docker compose (v2 plugin)
        try:
            result = subprocess.run(
                ["docker", "compose", "version", "--short"],
                capture_output=True,
                text=True,
                check=True,
            )
            version = result.stdout.strip()
            self._print_check(CheckStatus.PASS, f"Docker Compose (plugin): {version}")
            self.results.checks["docker_compose"] = {"version": version, "type": "plugin"}
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass

        # Check for docker-compose (standalone)
        if shutil.which("docker-compose"):
            try:
                result = subprocess.run(
                    ["docker-compose", "version", "--short"],
                    capture_output=True,
                    text=True,
                    check=True,
                )
                version = result.stdout.strip()
                self._print_check(CheckStatus.PASS, f"Docker Compose (standalone): {version}")
                self.results.checks["docker_compose"] = {"version": version, "type": "standalone"}
                return True
            except subprocess.CalledProcessError:
                pass

        self._print_check(CheckStatus.FAIL, "Docker Compose: not installed")
        self.results.errors.append("Docker Compose is required but not installed")
        self.results.fixes.append("Install Docker Compose: https://docs.docker.com/compose/install/")
        self.results.checks["docker_compose"] = {"installed": False}
        return False

    # =========================================================================
    # Network Connectivity Checks
    # =========================================================================

    def _check_network_server(self) -> None:
        """Check network connectivity for server mode."""
        self._print_section("Network Connectivity")

        self._check_dns()
        self._check_registry("docker.io")
        self._check_registry("ghcr.io", required=False)

    def _check_network_node(self) -> None:
        """Check network connectivity for node mode."""
        self._print_section("Network Connectivity")

        self._check_dns()
        self._check_config_server()
        self._check_registry("github.com", required=False)

    def _check_dns(self) -> bool:
        """Check DNS resolution."""
        self._verbose("Checking DNS resolution...")

        try:
            socket.gethostbyname("google.com")
            self._print_check(CheckStatus.PASS, "DNS: working")
            self.results.checks["dns"] = {"working": True}
            return True
        except socket.gaierror:
            self._print_check(CheckStatus.FAIL, "DNS: not working")
            self.results.errors.append("DNS resolution is not working")
            self.results.checks["dns"] = {"working": False}
            return False

    def _check_registry(self, registry: str, required: bool = True, timeout: int = 5) -> bool:
        """Check connectivity to a container registry."""
        self._verbose(f"Checking connectivity to {registry}...")

        try:
            result = subprocess.run(
                ["curl", "-sSf", "--connect-timeout", str(timeout), f"https://{registry}"],
                capture_output=True,
                timeout=timeout + 2,
            )
            if result.returncode == 0:
                self._print_check(CheckStatus.PASS, f"{registry}: reachable")
                self.results.checks[f"net_{registry.replace('.', '_')}"] = True
                return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            pass

        if required:
            self._print_check(CheckStatus.FAIL, f"{registry}: not reachable")
            self.results.errors.append(f"Cannot reach {registry} - check internet connectivity or firewall")
        else:
            self._print_check(CheckStatus.WARN, f"{registry}: not reachable (optional)")
            self.results.warnings.append(f"Cannot reach {registry}")

        self.results.checks[f"net_{registry.replace('.', '_')}"] = False
        return False

    def _check_config_server(self) -> bool:
        """Check Config Server connectivity."""
        if not self.config_server_url:
            self._verbose("Skipping Config Server check (no URL provided)")
            return True

        self._verbose("Checking Config Server connectivity...")

        health_url = f"{self.config_server_url.rstrip('/')}/api/v1/health"

        try:
            result = subprocess.run(
                ["curl", "-sSf", "--connect-timeout", "5", health_url],
                capture_output=True,
                timeout=10,
            )
            if result.returncode == 0:
                self._print_check(CheckStatus.PASS, f"Config Server: reachable ({self.config_server_url})")
                self.results.checks["config_server"] = {"reachable": True, "url": self.config_server_url}
                return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            pass

        self._print_check(CheckStatus.FAIL, f"Config Server: not reachable ({self.config_server_url})")
        self.results.errors.append(f"Cannot reach Config Server at {self.config_server_url}")
        self.results.checks["config_server"] = {"reachable": False, "url": self.config_server_url}
        return False

    # =========================================================================
    # Port Availability Checks
    # =========================================================================

    def _check_port(self, port: int) -> bool:
        """Check if a port is available."""
        self._verbose(f"Checking port {port}...")

        in_use = False
        process_info = ""

        # Try to bind to the port
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.bind(("127.0.0.1", port))
        except OSError:
            in_use = True
            process_info = self._get_port_process(port)

        self.results.checks[f"port_{port}"] = {
            "available": not in_use,
            "process": process_info if in_use else None,
        }

        if not in_use:
            self._print_check(CheckStatus.PASS, f"Port {port}: available")
            return True
        else:
            msg = f"Port {port}: in use"
            if process_info:
                msg += f" by {process_info}"
            self._print_check(CheckStatus.FAIL, msg)
            self.results.errors.append(f"Port {port} is already in use{f' by {process_info}' if process_info else ''}")
            return False

    def _get_port_process(self, port: int) -> str:
        """Get the process using a port."""
        try:
            if shutil.which("ss"):
                result = subprocess.run(
                    ["ss", "-tulnp"],
                    capture_output=True,
                    text=True,
                )
                for line in result.stdout.splitlines():
                    if f":{port} " in line:
                        # Extract process name
                        if "users:" in line:
                            start = line.find('("') + 2
                            end = line.find('"', start)
                            if start > 1 and end > start:
                                return line[start:end]
            elif shutil.which("netstat"):
                result = subprocess.run(
                    ["netstat", "-tulnp"],
                    capture_output=True,
                    text=True,
                )
                for line in result.stdout.splitlines():
                    if f":{port} " in line:
                        parts = line.split()
                        if len(parts) >= 7:
                            return parts[-1]
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
        return ""

    # =========================================================================
    # Permission Checks
    # =========================================================================

    def _check_permissions_server(self) -> None:
        """Check permissions for server mode."""
        self._print_section("Permissions")

        self._check_root_or_sudo()
        self._check_docker_socket()

    def _check_permissions_node(self) -> None:
        """Check permissions for node mode."""
        self._print_section("Permissions")

        self._check_root_or_sudo()

        # Check write access to systemd directory
        systemd_dir = Path("/etc/systemd/system")
        if systemd_dir.exists():
            if os.access(systemd_dir, os.W_OK):
                self._print_check(CheckStatus.PASS, "/etc/systemd/system: writable")
            else:
                # Try with sudo
                try:
                    result = subprocess.run(
                        ["sudo", "test", "-w", str(systemd_dir)],
                        capture_output=True,
                        timeout=5,
                    )
                    if result.returncode == 0:
                        self._print_check(CheckStatus.PASS, "/etc/systemd/system: writable (with sudo)")
                    else:
                        self._print_check(CheckStatus.FAIL, "/etc/systemd/system: not writable")
                        self.results.errors.append("/etc/systemd/system is not writable")
                except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
                    self._print_check(CheckStatus.FAIL, "/etc/systemd/system: not writable")
                    self.results.errors.append("/etc/systemd/system is not writable")

    def _check_root_or_sudo(self) -> bool:
        """Check for root or sudo access."""
        self._verbose("Checking root/sudo access...")

        is_root = os.geteuid() == 0

        if is_root:
            self._print_check(CheckStatus.PASS, "Running as root")
            self.results.checks["permissions"] = {"is_root": True, "has_sudo": True}
            return True

        # Check for passwordless sudo
        try:
            result = subprocess.run(
                ["sudo", "-n", "true"],
                capture_output=True,
                timeout=5,
            )
            if result.returncode == 0:
                self._print_check(CheckStatus.PASS, "sudo: available (passwordless)")
                self.results.checks["permissions"] = {"is_root": False, "has_sudo": True}
                return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            pass

        # Check for sudo with password
        try:
            result = subprocess.run(
                ["sudo", "-v"],
                capture_output=True,
                timeout=5,
            )
            if result.returncode == 0:
                self._print_check(CheckStatus.PASS, "sudo: available")
                self.results.checks["permissions"] = {"is_root": False, "has_sudo": True}
                return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            pass

        self._print_check(CheckStatus.FAIL, "sudo: not available")
        self.results.errors.append("Root or sudo access is required for installation")
        self.results.checks["permissions"] = {"is_root": False, "has_sudo": False}
        return False

    def _check_docker_socket(self) -> bool:
        """Check Docker socket accessibility."""
        self._verbose("Checking Docker socket access...")

        socket_path = Path("/var/run/docker.sock")

        if not socket_path.exists():
            self._print_check(CheckStatus.WARN, "Docker socket: not found")
            self.results.warnings.append("Docker socket not found at /var/run/docker.sock")
            return True

        try:
            subprocess.run(["docker", "info"], capture_output=True, check=True, timeout=10)
            self._print_check(CheckStatus.PASS, "Docker socket: accessible")
            self.results.checks["docker_socket"] = {"accessible": True}
            return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            self._print_check(CheckStatus.FAIL, "Docker socket: not accessible")
            self.results.errors.append(
                "Cannot access Docker socket. Add user to docker group: sudo usermod -aG docker $USER"
            )
            self.results.checks["docker_socket"] = {"accessible": False}

            if self.fix_mode:
                self._info("Attempting to add current user to docker group...")
                self._try_add_docker_group()
            return False

    # =========================================================================
    # Hardware Detection
    # =========================================================================

    def _check_hardware(self) -> None:
        """Detect hardware (GPUs, InfiniBand)."""
        self._print_section("Hardware Detection")

        hardware = {}

        # NVIDIA GPU
        if shutil.which("nvidia-smi"):
            try:
                # Get GPU count
                result = subprocess.run(
                    ["nvidia-smi", "--query-gpu=count", "--format=csv,noheader"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                gpu_count = result.stdout.strip().split("\n")[0] if result.returncode == 0 else "0"

                # Get GPU model
                result = subprocess.run(
                    ["nvidia-smi", "--query-gpu=name", "--format=csv,noheader"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                gpu_model = result.stdout.strip().split("\n")[0] if result.returncode == 0 else "unknown"

                # Get driver version
                result = subprocess.run(
                    ["nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                driver_version = result.stdout.strip().split("\n")[0] if result.returncode == 0 else "unknown"

                self._print_check(CheckStatus.INFO, f"NVIDIA GPU detected: {gpu_count}x {gpu_model}")
                self._print_check(CheckStatus.INFO, f"NVIDIA Driver: {driver_version}")

                hardware["gpu"] = {
                    "vendor": "nvidia",
                    "count": gpu_count,
                    "model": gpu_model,
                    "driver": driver_version,
                }
            except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
                pass
        else:
            self._verbose("No NVIDIA GPU detected")
            hardware["gpu"] = {"vendor": "none"}

        # AMD GPU
        if shutil.which("rocm-smi"):
            self._print_check(CheckStatus.INFO, "AMD GPU detected (ROCm available)")
            hardware["amd_gpu"] = True

        # InfiniBand
        if shutil.which("ibstat"):
            try:
                result = subprocess.run(
                    ["ibstat", "-l"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                ib_devices = len(result.stdout.strip().split("\n")) if result.returncode == 0 else 0
                if ib_devices > 0:
                    self._print_check(CheckStatus.INFO, f"InfiniBand: {ib_devices} device(s) detected")
                    hardware["infiniband"] = ib_devices
            except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
                pass

        self.results.checks["hardware"] = hardware

    # =========================================================================
    # Helper Methods
    # =========================================================================

    def _parse_os_release(self) -> dict[str, str]:
        """Parse /etc/os-release file."""
        os_release = {}
        os_release_path = Path("/etc/os-release")

        if os_release_path.exists():
            with open(os_release_path) as f:
                for line in f:
                    line = line.strip()
                    if "=" in line:
                        key, value = line.split("=", 1)
                        os_release[key] = value.strip('"')

        return os_release

    def _get_total_ram_gb(self) -> int:
        """Get total RAM in GB."""
        system = platform.system()

        if system == "Darwin":
            try:
                result = subprocess.run(
                    ["sysctl", "-n", "hw.memsize"],
                    capture_output=True,
                    text=True,
                    check=True,
                )
                return int(result.stdout.strip()) // (1024**3)
            except (subprocess.CalledProcessError, ValueError):
                return 0

        # Linux
        try:
            with open("/proc/meminfo") as f:
                for line in f:
                    if line.startswith("MemTotal:"):
                        kb = int(line.split()[1])
                        return kb // (1024**2)
        except (FileNotFoundError, ValueError, IndexError):
            pass

        return 0

    def _get_available_disk_gb(self) -> int:
        """Get available disk space in GB."""
        try:
            stat = os.statvfs("/")
            return (stat.f_bavail * stat.f_frsize) // (1024**3)
        except OSError:
            return 0

    def _get_command_version(self, cmd: str) -> str:
        """Get version string for a command."""
        try:
            if cmd == "docker":
                result = subprocess.run(
                    ["docker", "version", "--format", "{{.Server.Version}}"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                if result.returncode == 0 and result.stdout.strip():
                    return result.stdout.strip()
                # Fallback
                result = subprocess.run(
                    ["docker", "--version"],
                    capture_output=True,
                    text=True,
                )
                return result.stdout.strip()
            else:
                result = subprocess.run(
                    [cmd, "--version"],
                    capture_output=True,
                    text=True,
                    timeout=5,
                )
                return result.stdout.strip().split("\n")[0] if result.returncode == 0 else "installed"
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            return "installed"

    def _try_install_docker(self) -> None:
        """Attempt to install Docker."""
        if shutil.which("apt-get"):
            subprocess.run(["apt-get", "update"], capture_output=True)
            subprocess.run(["apt-get", "install", "-y", "docker.io"], capture_output=True)
        elif shutil.which("yum"):
            subprocess.run(["yum", "install", "-y", "docker"], capture_output=True)

    def _try_start_docker(self) -> None:
        """Attempt to start Docker daemon."""
        if shutil.which("systemctl"):
            subprocess.run(["systemctl", "start", "docker"], capture_output=True)
        elif shutil.which("service"):
            subprocess.run(["service", "docker", "start"], capture_output=True)

    def _try_add_docker_group(self) -> None:
        """Attempt to add current user to docker group."""
        user = os.environ.get("USER", "")
        if user:
            subprocess.run(["sudo", "usermod", "-aG", "docker", user], capture_output=True)
            self._warn("Please log out and back in for group changes to take effect")

    # =========================================================================
    # Output Methods
    # =========================================================================

    def _print_header(self) -> None:
        """Print header banner."""
        if self.json_output or self.quiet_mode:
            return

        print()
        print(f"{Colors.BOLD}AAMI Preflight Check v{VERSION}{Colors.NC}")
        print("=========================")
        print()
        print(f"Mode: {Colors.CYAN}{self.mode.value.title()} Installation{Colors.NC}")

    def _print_section(self, title: str) -> None:
        """Print section header."""
        if self.json_output or self.quiet_mode:
            return
        print()
        print(f"{Colors.BOLD}{title}{Colors.NC}")

    def _print_check(self, status: CheckStatus, message: str) -> None:
        """Print check result."""
        if self.json_output:
            return

        icons = {
            CheckStatus.PASS: f"{Colors.GREEN}[✓]{Colors.NC}",
            CheckStatus.FAIL: f"{Colors.RED}[✗]{Colors.NC}",
            CheckStatus.WARN: f"{Colors.YELLOW}[!]{Colors.NC}",
            CheckStatus.INFO: f"{Colors.BLUE}[i]{Colors.NC}",
        }
        print(f"  {icons[status]} {message}")

    def _print_summary(self) -> None:
        """Print summary of results."""
        if self.json_output:
            return

        print()
        print("━" * 40)

        if self.results.passed:
            print(f"{Colors.GREEN}Result: All checks passed!{Colors.NC}")
        else:
            print(f"{Colors.RED}Result: {len(self.results.errors)} issue(s) found{Colors.NC}")
            print()
            print("ERRORS:")
            for err in self.results.errors:
                print(f"  {Colors.RED}[✗]{Colors.NC} {err}")

        if self.results.warnings:
            print()
            print("WARNINGS:")
            for warn in self.results.warnings:
                print(f"  {Colors.YELLOW}[!]{Colors.NC} {warn}")

        if self.results.fixes and not self.fix_mode:
            print()
            print("SUGGESTED FIXES:")
            for fix in self.results.fixes:
                print(f"  {Colors.BLUE}→{Colors.NC} {fix}")
            print()
            print("Run with --fix to attempt automatic fixes.")

        print()

    def _output_json(self) -> None:
        """Output results as JSON."""
        output = {
            "version": self.results.version,
            "mode": self.results.mode,
            "timestamp": self.results.timestamp,
            "passed": self.results.passed,
            "checks": self.results.checks,
            "errors": self.results.errors,
            "warnings": self.results.warnings,
        }
        print(json.dumps(output, indent=2))

    def _info(self, message: str) -> None:
        """Print info message."""
        if not self.quiet_mode and not self.json_output:
            print(f"{Colors.GREEN}[INFO]{Colors.NC} {message}")

    def _warn(self, message: str) -> None:
        """Print warning message."""
        if not self.json_output:
            print(f"{Colors.YELLOW}[WARN]{Colors.NC} {message}")
        self.results.warnings.append(message)

    def _verbose(self, message: str) -> None:
        """Print verbose message."""
        if self.verbose_mode and not self.json_output:
            print(f"{Colors.CYAN}[DEBUG]{Colors.NC} {message}")


def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="AAMI Preflight Check - Validates system requirements before installation.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # Basic server check
    %(prog)s --mode server

    # Node check with connectivity test
    %(prog)s --mode node --server https://config.example.com

    # Auto-fix issues
    %(prog)s --mode server --fix

    # JSON output for CI/CD
    %(prog)s --mode node --json

Exit Codes:
    0 - All checks passed
    1 - One or more checks failed
    2 - Invalid arguments
        """,
    )

    parser.add_argument(
        "--mode",
        choices=["server", "node"],
        help="Check mode: 'server' or 'node' (default: auto-detect)",
    )
    parser.add_argument(
        "--server",
        metavar="URL",
        help="Config Server URL (for node mode connectivity check)",
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Attempt automatic fixes for issues",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Output results in JSON format",
    )
    parser.add_argument(
        "--quiet",
        action="store_true",
        help="Only show errors",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Show detailed check information",
    )
    parser.add_argument(
        "-V",
        "--version",
        action="version",
        version=f"%(prog)s {VERSION}",
    )

    args = parser.parse_args()

    checker = PreflightChecker(
        mode=args.mode,
        config_server_url=args.server or "",
        fix_mode=args.fix,
        json_output=args.json,
        quiet_mode=args.quiet,
        verbose_mode=args.verbose,
    )

    sys.exit(checker.run())


if __name__ == "__main__":
    main()
