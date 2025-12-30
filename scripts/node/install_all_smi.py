#!/usr/bin/env python3
"""
Install all-smi Multi-Vendor AI Accelerator Metrics Exporter

This script downloads and installs all-smi as a systemd service
on Linux systems. all-smi provides unified Prometheus metrics for
multiple AI accelerator vendors (NVIDIA, AMD, Intel, etc.)

Usage:
    sudo ./install_all_smi.py [OPTIONS]

Options:
    -v, --version VERSION    all-smi version (default: 0.5.0)
    -p, --port PORT          Listen port (default: 9401)
    -h, --help               Show this help message

Environment Variables:
    ALL_SMI_VERSION   - Version to install
    ALL_SMI_PORT      - Port to listen on

Reference: https://github.com/lablup/all-smi
"""

import argparse
import grp
import os
import pwd
import shutil
import subprocess
import sys
import time
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

VERSION = "1.0.0"

# Default configuration
DEFAULT_ALL_SMI_VERSION = "0.5.0"
DEFAULT_ALL_SMI_PORT = 9401
SERVICE_USER = "all_smi"
SERVICE_FILE = Path("/etc/systemd/system/all-smi.service")

# ANSI colors
RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
CYAN = "\033[0;36m"
BOLD = "\033[1m"
NC = "\033[0m"  # No Color


@dataclass
class Config:
    """Installation configuration."""
    version: str
    port: int
    service_user: str = SERVICE_USER


class AllSmiInstaller:
    """all-smi installer class."""

    def __init__(self, config: Config) -> None:
        self._config = config
        self._all_smi_bin: Optional[str] = None

    def _info(self, message: str) -> None:
        """Print info message."""
        print(f"{GREEN}[INFO]{NC} {message}")

    def _warn(self, message: str) -> None:
        """Print warning message."""
        print(f"{YELLOW}[WARN]{NC} {message}")

    def _error(self, message: str) -> None:
        """Print error message."""
        print(f"{RED}[ERROR]{NC} {message}", file=sys.stderr)

    def _run_command(
        self,
        cmd: list[str],
        check: bool = True,
        capture_output: bool = False,
    ) -> subprocess.CompletedProcess:
        """Run a shell command."""
        return subprocess.run(
            cmd,
            check=check,
            capture_output=capture_output,
            text=True,
        )

    def _check_root(self) -> bool:
        """Check if running as root."""
        if os.geteuid() != 0:
            self._error("This script must be run as root")
            return False
        return True

    def _detect_architecture(self) -> Optional[str]:
        """Detect system architecture."""
        import platform
        machine = platform.machine()
        arch_map = {
            "x86_64": "amd64",
            "aarch64": "arm64",
            "arm64": "arm64",
        }
        arch = arch_map.get(machine)
        if not arch:
            self._error(f"Unsupported architecture: {machine}")
            return None
        return arch

    def _check_python_pip(self) -> bool:
        """Check if Python3 and pip are available."""
        if not shutil.which("python3"):
            self._error("Python3 is required but not installed")
            return False

        if not shutil.which("pip3"):
            self._error("pip3 is required but not installed")
            return False

        return True

    def _install_all_smi(self) -> bool:
        """Install all-smi via pip."""
        self._info("Installing all-smi via pip...")

        # Try specific version first
        try:
            self._run_command(
                ["pip3", "install", f"all-smi=={self._config.version}", "--quiet"],
                capture_output=True,
            )
            return True
        except subprocess.CalledProcessError:
            self._warn("Specific version not found, installing latest version...")

        # Fall back to latest version
        try:
            self._run_command(
                ["pip3", "install", "all-smi", "--quiet"],
                capture_output=True,
            )
            return True
        except subprocess.CalledProcessError:
            self._error("Failed to install all-smi")
            return False

    def _find_binary(self) -> bool:
        """Find the all-smi binary location."""
        # Try which first
        binary = shutil.which("all-smi")
        if binary:
            self._all_smi_bin = binary
            self._info(f"all-smi binary found at: {self._all_smi_bin}")
            return True

        # Try common pip install locations
        common_paths = [
            "/usr/local/bin/all-smi",
            Path.home() / ".local/bin/all-smi",
            "/usr/bin/all-smi",
        ]

        for path in common_paths:
            path = Path(path)
            if path.exists():
                self._all_smi_bin = str(path)
                self._info(f"all-smi binary found at: {self._all_smi_bin}")
                return True

        self._error("all-smi binary not found after installation")
        return False

    def _create_service_user(self) -> bool:
        """Create service user if it doesn't exist."""
        try:
            pwd.getpwnam(self._config.service_user)
            self._info(f"Service user {self._config.service_user} already exists")
        except KeyError:
            self._info(f"Creating service user: {self._config.service_user}")
            try:
                self._run_command([
                    "useradd",
                    "--no-create-home",
                    "--shell", "/bin/false",
                    "--system",
                    self._config.service_user,
                ])
            except subprocess.CalledProcessError as e:
                self._error(f"Failed to create service user: {e}")
                return False

        return True

    def _add_user_to_groups(self) -> None:
        """Add service user to video and render groups for GPU access."""
        groups_to_add = []

        # Check if video group exists
        try:
            grp.getgrnam("video")
            groups_to_add.append("video")
        except KeyError:
            pass

        # Check if render group exists
        try:
            grp.getgrnam("render")
            groups_to_add.append("render")
        except KeyError:
            pass

        for group in groups_to_add:
            self._info(f"Adding {self._config.service_user} to {group} group for GPU access")
            try:
                self._run_command([
                    "usermod", "-aG", group, self._config.service_user,
                ])
            except subprocess.CalledProcessError:
                self._warn(f"Failed to add user to {group} group")

    def _create_systemd_service(self) -> bool:
        """Create systemd service file."""
        self._info("Creating systemd service...")

        service_content = f"""[Unit]
Description=all-smi Multi-Vendor AI Accelerator Metrics Exporter
Documentation=https://github.com/lablup/all-smi
Wants=network-online.target
After=network-online.target

[Service]
User={self._config.service_user}
Group={self._config.service_user}
Type=simple
ExecStart={self._all_smi_bin} serve --port {self._config.port}
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
ProtectHome=true
PrivateTmp=true

# Allow GPU device access
SupplementaryGroups=video render

[Install]
WantedBy=multi-user.target
"""

        try:
            SERVICE_FILE.write_text(service_content)
            return True
        except OSError as e:
            self._error(f"Failed to create service file: {e}")
            return False

    def _enable_and_start_service(self) -> bool:
        """Reload systemd, enable and start the service."""
        self._info("Reloading systemd daemon...")
        try:
            self._run_command(["systemctl", "daemon-reload"])
        except subprocess.CalledProcessError as e:
            self._error(f"Failed to reload systemd: {e}")
            return False

        self._info("Enabling all-smi service...")
        try:
            self._run_command(["systemctl", "enable", "all-smi"])
        except subprocess.CalledProcessError as e:
            self._error(f"Failed to enable service: {e}")
            return False

        self._info("Starting all-smi service...")
        try:
            self._run_command(["systemctl", "start", "all-smi"])
        except subprocess.CalledProcessError as e:
            self._error(f"Failed to start service: {e}")
            return False

        return True

    def _verify_installation(self) -> bool:
        """Verify the installation is working."""
        self._info("Verifying installation...")
        time.sleep(2)

        # Check service status
        result = self._run_command(
            ["systemctl", "is-active", "--quiet", "all-smi"],
            check=False,
        )

        if result.returncode != 0:
            self._error("all-smi service failed to start")
            self._error("Check logs with: journalctl -u all-smi -f")
            return False

        self._info("all-smi installed and running successfully!")
        self._info(f"Metrics available at: http://localhost:{self._config.port}/metrics")

        # Test metrics endpoint
        try:
            url = f"http://localhost:{self._config.port}/metrics"
            with urllib.request.urlopen(url, timeout=5) as response:
                if response.status == 200:
                    self._info("Metrics endpoint is responding")
        except Exception:
            self._warn("Metrics endpoint is not responding yet. Service may still be initializing.")

        return True

    def _detect_accelerators(self) -> None:
        """Detect available AI accelerators."""
        self._info("Detecting available AI accelerators...")
        detected = []

        # Check NVIDIA GPUs
        if shutil.which("nvidia-smi"):
            try:
                result = self._run_command(
                    ["nvidia-smi", "--query-gpu=count", "--format=csv,noheader,nounits"],
                    capture_output=True,
                    check=False,
                )
                if result.returncode == 0:
                    count = result.stdout.strip().split("\n")[0]
                    if count and int(count) > 0:
                        detected.append(f"NVIDIA GPU ({count}x)")
            except (subprocess.CalledProcessError, ValueError):
                pass

        # Check AMD GPUs
        if shutil.which("rocm-smi"):
            try:
                result = self._run_command(
                    ["rocm-smi", "--showproductname"],
                    capture_output=True,
                    check=False,
                )
                if result.returncode == 0:
                    count = result.stdout.count("GPU")
                    if count > 0:
                        detected.append(f"AMD GPU ({count}x)")
            except subprocess.CalledProcessError:
                pass

        if detected:
            self._info(f"Detected accelerators: {', '.join(detected)}")
        else:
            self._warn("No AI accelerators detected. all-smi will report no metrics until accelerators are available.")

    def _print_summary(self) -> None:
        """Print installation summary."""
        print(f"""
{GREEN}Installation complete!{NC}

all-smi is now monitoring AI accelerators on this system.

Port: {self._config.port}
Binary: {self._all_smi_bin}
Service: all-smi.service

Next steps:
1. Register this exporter in AAMI Config Server
2. Add firewall rule if needed:
   sudo ufw allow {self._config.port}/tcp
3. Verify metrics:
   curl http://localhost:{self._config.port}/metrics

Service commands:
  Status:  sudo systemctl status all-smi
  Stop:    sudo systemctl stop all-smi
  Start:   sudo systemctl start all-smi
  Restart: sudo systemctl restart all-smi
  Logs:    sudo journalctl -u all-smi -f
""")

    def install(self) -> int:
        """Run the full installation process."""
        # Step 1: Check root
        if not self._check_root():
            return 1

        # Step 2: Detect architecture
        arch = self._detect_architecture()
        if not arch:
            return 1

        self._info(f"Installing all-smi v{self._config.version} for {arch}")

        # Step 3: Check Python/pip
        if not self._check_python_pip():
            return 1

        # Step 4: Install all-smi
        if not self._install_all_smi():
            return 1

        # Step 5: Find binary
        if not self._find_binary():
            return 1

        # Step 6: Create service user
        if not self._create_service_user():
            return 1

        # Step 7: Add user to groups
        self._add_user_to_groups()

        # Step 8: Create systemd service
        if not self._create_systemd_service():
            return 1

        # Step 9: Enable and start service
        if not self._enable_and_start_service():
            return 1

        # Step 10: Verify installation
        if not self._verify_installation():
            return 1

        # Step 11: Detect accelerators
        self._detect_accelerators()

        # Step 12: Print summary
        self._print_summary()

        return 0


def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Install all-smi Multi-Vendor AI Accelerator Metrics Exporter",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
all-smi supports:
  - NVIDIA GPUs (CUDA)
  - AMD GPUs (ROCm)
  - Intel Gaudi NPUs
  - Google Cloud TPUs
  - Apple Silicon GPUs
  - Tenstorrent, Rebellions, Furiosa NPUs

Environment Variables:
    ALL_SMI_VERSION   - Version to install
    ALL_SMI_PORT      - Port to listen on

Example:
    sudo %(prog)s
    sudo %(prog)s --version 0.5.0 --port 9401
        """,
    )

    parser.add_argument(
        "-v", "--version",
        metavar="VERSION",
        default=os.environ.get("ALL_SMI_VERSION", DEFAULT_ALL_SMI_VERSION),
        help=f"all-smi version (default: {DEFAULT_ALL_SMI_VERSION})",
    )
    parser.add_argument(
        "-p", "--port",
        type=int,
        default=int(os.environ.get("ALL_SMI_PORT", DEFAULT_ALL_SMI_PORT)),
        help=f"Listen port (default: {DEFAULT_ALL_SMI_PORT})",
    )
    parser.add_argument(
        "-V", "--script-version",
        action="version",
        version=f"%(prog)s {VERSION}",
    )

    args = parser.parse_args()

    config = Config(
        version=args.version,
        port=args.port,
    )

    installer = AllSmiInstaller(config)
    sys.exit(installer.install())


if __name__ == "__main__":
    main()
