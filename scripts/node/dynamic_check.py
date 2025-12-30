#!/usr/bin/env python3
"""
Dynamic Check Runner for AAMI Monitoring

This script fetches effective checks from the AAMI Config Server,
executes them, and outputs results to Node Exporter's textfile collector.

Usage:
    ./dynamic_check.py [OPTIONS]

Options:
    -c, --config-server URL  Config Server URL (default: from /etc/aami/config)
    -h, --hostname NAME      Override hostname (default: system hostname)
    -d, --debug              Enable debug logging
    --help                   Show this help message

Environment Variables:
    AAMI_CONFIG_SERVER_URL   - Config Server URL
    AAMI_HOSTNAME            - Override hostname
    AAMI_DEBUG               - Enable debug logging (1=on, 0=off)
"""

import argparse
import json
import logging
import os
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Optional

VERSION = "1.0.0"

# Default paths
DEFAULT_TEXTFILE_DIR = "/var/lib/node_exporter/textfile_collector"
DEFAULT_CHECK_SCRIPTS_DIR = "/usr/local/lib/aami/checks"
DEFAULT_CONFIG_FILE = "/etc/aami/config"
DEFAULT_LOG_FILE = "/var/log/aami/dynamic-check.log"


@dataclass
class CheckInfo:
    """Information about a single check from Config Server."""
    name: str
    script_content: str
    script_hash: str
    config: dict


@dataclass
class CheckResult:
    """Result of executing a check."""
    name: str
    success: bool
    output: str
    error: Optional[str] = None


class DynamicCheckRunner:
    """Main dynamic check runner."""

    def __init__(
        self,
        config_server_url: str = "",
        hostname: str = "",
        debug: bool = False,
        textfile_dir: str = DEFAULT_TEXTFILE_DIR,
        check_scripts_dir: str = DEFAULT_CHECK_SCRIPTS_DIR,
        config_file: str = DEFAULT_CONFIG_FILE,
        log_file: str = DEFAULT_LOG_FILE,
    ) -> None:
        self.config_server_url = config_server_url
        self.hostname = hostname or socket.gethostname()
        self.debug = debug
        self.textfile_dir = Path(textfile_dir)
        self.check_scripts_dir = Path(check_scripts_dir)
        self.config_file = Path(config_file)
        self.log_file = Path(log_file)

        self._setup_logging()
        self._load_config()
        self._ensure_directories()

    def _setup_logging(self) -> None:
        """Setup logging configuration."""
        log_level = logging.DEBUG if self.debug else logging.INFO
        log_format = "%(asctime)s [%(levelname)s] %(message)s"

        handlers: list[logging.Handler] = []

        # Try to setup file logging
        try:
            self.log_file.parent.mkdir(parents=True, exist_ok=True)
            handlers.append(logging.FileHandler(self.log_file))
        except PermissionError:
            # Fall back to stderr if we can't write to log file
            pass

        # Add console handler for debug mode
        if self.debug:
            handlers.append(logging.StreamHandler())

        # Ensure at least one handler exists
        if not handlers:
            handlers.append(logging.StreamHandler())

        logging.basicConfig(
            level=log_level,
            format=log_format,
            handlers=handlers,
        )
        self.logger = logging.getLogger(__name__)

    def _load_config(self) -> None:
        """Load configuration from config file if URL not provided."""
        if self.config_server_url:
            return

        if self.config_file.exists():
            try:
                with open(self.config_file) as f:
                    for line in f:
                        line = line.strip()
                        if line.startswith("AAMI_CONFIG_SERVER_URL="):
                            value = line.split("=", 1)[1].strip('"').strip("'")
                            self.config_server_url = value
                            break
            except OSError as e:
                self.logger.warning(f"Could not read config file: {e}")

        if not self.config_server_url:
            self.logger.error("Config Server URL not configured")
            self.logger.error("Set via: AAMI_CONFIG_SERVER_URL env, config file, or --config-server flag")

    def _ensure_directories(self) -> None:
        """Create required directories if they don't exist."""
        self.textfile_dir.mkdir(parents=True, exist_ok=True)
        self.check_scripts_dir.mkdir(parents=True, exist_ok=True)

    def run(self) -> int:
        """Run dynamic checks and return exit code."""
        if not self.config_server_url:
            self._write_status_metrics(success=False, checks_total=0, checks_success=0, checks_failed=0, duration=0)
            return 1

        start_time = time.time()
        self.logger.info(f"Starting dynamic check run for hostname: {self.hostname}")
        self.logger.debug(f"Config Server: {self.config_server_url}")
        self.logger.debug(f"Textfile Directory: {self.textfile_dir}")
        self.logger.debug(f"Check Scripts Directory: {self.check_scripts_dir}")

        # Fetch effective checks
        checks = self._fetch_effective_checks()
        if checks is None:
            self._write_status_metrics(
                success=False,
                checks_total=0,
                checks_success=0,
                checks_failed=0,
                duration=int(time.time() - start_time),
            )
            return 1

        # Execute checks
        checks_total = len(checks)
        checks_success = 0
        checks_failed = 0

        for check in checks:
            self.logger.info(f"Processing check: {check.name}")

            # Save script
            script_path = self._save_check_script(check)

            # Execute check
            result = self._execute_check(check.name, script_path, check.config)

            if result.success:
                checks_success += 1
            else:
                checks_failed += 1

        # Write status metrics
        duration = int(time.time() - start_time)
        self._write_status_metrics(
            success=True,
            checks_total=checks_total,
            checks_success=checks_success,
            checks_failed=checks_failed,
            duration=duration,
        )

        self.logger.info(
            f"Check run completed: total={checks_total}, success={checks_success}, "
            f"failed={checks_failed}, duration={duration}s"
        )

        return 0

    def _fetch_effective_checks(self) -> Optional[list[CheckInfo]]:
        """Fetch effective checks from Config Server."""
        url = f"{self.config_server_url.rstrip('/')}/api/v1/checks/target/hostname/{self.hostname}"
        self.logger.debug(f"Fetching effective checks from: {url}")

        try:
            request = urllib.request.Request(
                url,
                headers={"Accept": "application/json"},
            )
            with urllib.request.urlopen(request, timeout=30) as response:
                data = json.loads(response.read().decode("utf-8"))

            checks = []
            for item in data:
                check = CheckInfo(
                    name=item.get("name", ""),
                    script_content=item.get("script_content", ""),
                    script_hash=item.get("script_hash", ""),
                    config=item.get("config") or {},
                )
                checks.append(check)

            self.logger.debug(f"Received {len(checks)} checks")
            return checks

        except urllib.error.URLError as e:
            self.logger.error(f"Failed to fetch effective checks: {e}")
            return None
        except json.JSONDecodeError as e:
            self.logger.error(f"Failed to parse response JSON: {e}")
            return None
        except Exception as e:
            self.logger.error(f"Unexpected error fetching checks: {e}")
            return None

    def _save_check_script(self, check: CheckInfo) -> Path:
        """Save check script with hash-based versioning."""
        script_dir = self.check_scripts_dir / check.name
        script_dir.mkdir(parents=True, exist_ok=True)

        script_file = script_dir / f"{check.name}_{check.script_hash}.sh"
        current_link = script_dir / "current.sh"

        # Check if script already exists with this hash
        if script_file.exists():
            self.logger.debug(f"Check script already exists: {script_file}")
        else:
            self.logger.info(f"Saving new check script: {check.name} (hash: {check.script_hash[:8]})")
            script_file.write_text(check.script_content)
            script_file.chmod(0o755)

        # Update symlink to current version
        if current_link.is_symlink():
            current_link.unlink()
        elif current_link.exists():
            current_link.unlink()
        current_link.symlink_to(script_file)

        return current_link

    def _execute_check(self, check_name: str, script_path: Path, config: dict) -> CheckResult:
        """Execute a single check."""
        output_file = self.textfile_dir / f"{check_name}.prom.tmp"
        final_file = self.textfile_dir / f"{check_name}.prom"

        self.logger.debug(f"Executing check: {check_name}")
        self.logger.debug(f"Script: {script_path}")
        self.logger.debug(f"Config: {config}")

        try:
            # Execute check script with config as stdin
            config_json = json.dumps(config)
            result = subprocess.run(
                [str(script_path)],
                input=config_json,
                capture_output=True,
                text=True,
                timeout=30,
            )

            if result.returncode == 0:
                # Write output to temp file, then atomically move
                output_file.write_text(result.stdout)
                output_file.rename(final_file)
                self.logger.info(f"Check completed successfully: {check_name}")
                return CheckResult(name=check_name, success=True, output=result.stdout)
            else:
                self.logger.error(f"Check failed: {check_name} (exit code: {result.returncode})")
                if result.stderr:
                    self.logger.error(f"Check stderr: {result.stderr}")

                # Write error metric
                error_output = self._generate_error_metric(check_name)
                output_file.write_text(error_output)
                output_file.rename(final_file)

                return CheckResult(
                    name=check_name,
                    success=False,
                    output=result.stdout,
                    error=result.stderr,
                )

        except subprocess.TimeoutExpired:
            self.logger.error(f"Check timed out: {check_name}")
            error_output = self._generate_error_metric(check_name)
            output_file.write_text(error_output)
            output_file.rename(final_file)
            return CheckResult(name=check_name, success=False, output="", error="Timeout")

        except Exception as e:
            self.logger.error(f"Check execution error: {check_name} - {e}")
            error_output = self._generate_error_metric(check_name)
            output_file.write_text(error_output)
            output_file.rename(final_file)
            return CheckResult(name=check_name, success=False, output="", error=str(e))

    def _generate_error_metric(self, check_name: str) -> str:
        """Generate error metric for a failed check."""
        return f"""# HELP aami_check_error Check execution error (1=error)
# TYPE aami_check_error gauge
aami_check_error{{check="{check_name}"}} 1
"""

    def _write_status_metrics(
        self,
        success: bool,
        checks_total: int,
        checks_success: int,
        checks_failed: int,
        duration: int,
    ) -> None:
        """Write overall status metrics."""
        timestamp = int(time.time())
        status_value = 1 if success else 0

        metrics = f"""# HELP aami_check_fetch_status Check fetch status (1=success, 0=failed)
# TYPE aami_check_fetch_status gauge
aami_check_fetch_status {status_value}

# HELP aami_check_fetch_timestamp_seconds Last check fetch timestamp
# TYPE aami_check_fetch_timestamp_seconds gauge
aami_check_fetch_timestamp_seconds {timestamp}

# HELP aami_check_execution_duration_seconds Check execution duration
# TYPE aami_check_execution_duration_seconds gauge
aami_check_execution_duration_seconds {duration}

# HELP aami_checks_total Total number of checks configured
# TYPE aami_checks_total gauge
aami_checks_total {checks_total}

# HELP aami_checks_success Number of successful checks
# TYPE aami_checks_success gauge
aami_checks_success {checks_success}

# HELP aami_checks_failed Number of failed checks
# TYPE aami_checks_failed gauge
aami_checks_failed {checks_failed}
"""

        status_file = self.textfile_dir / "aami_status.prom"
        temp_file = self.textfile_dir / "aami_status.prom.tmp"

        temp_file.write_text(metrics)
        temp_file.rename(status_file)


def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Dynamic Check Runner for AAMI Monitoring",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Environment Variables:
    AAMI_CONFIG_SERVER_URL   - Config Server URL
    AAMI_HOSTNAME            - Override hostname
    AAMI_DEBUG               - Enable debug logging (1=on, 0=off)

Example:
    %(prog)s --config-server http://config-server:8080
    %(prog)s --debug
        """,
    )

    parser.add_argument(
        "-c", "--config-server",
        metavar="URL",
        default=os.environ.get("AAMI_CONFIG_SERVER_URL", ""),
        help="Config Server URL (default: from /etc/aami/config)",
    )
    parser.add_argument(
        "--hostname",
        default=os.environ.get("AAMI_HOSTNAME", ""),
        help=f"Override hostname (default: {socket.gethostname()})",
    )
    parser.add_argument(
        "-d", "--debug",
        action="store_true",
        default=os.environ.get("AAMI_DEBUG", "0") == "1",
        help="Enable debug logging",
    )
    parser.add_argument(
        "--textfile-dir",
        default=os.environ.get("TEXTFILE_DIR", DEFAULT_TEXTFILE_DIR),
        help=f"Textfile collector directory (default: {DEFAULT_TEXTFILE_DIR})",
    )
    parser.add_argument(
        "--check-scripts-dir",
        default=os.environ.get("CHECK_SCRIPTS_DIR", DEFAULT_CHECK_SCRIPTS_DIR),
        help=f"Check scripts directory (default: {DEFAULT_CHECK_SCRIPTS_DIR})",
    )
    parser.add_argument(
        "-V", "--version",
        action="version",
        version=f"%(prog)s {VERSION}",
    )

    args = parser.parse_args()

    runner = DynamicCheckRunner(
        config_server_url=args.config_server,
        hostname=args.hostname,
        debug=args.debug,
        textfile_dir=args.textfile_dir,
        check_scripts_dir=args.check_scripts_dir,
    )

    sys.exit(runner.run())


if __name__ == "__main__":
    main()
