package installer

import (
	"context"
	"fmt"
	"strings"

	"github.com/fregataa/aami/internal/ssh"
)

// ExporterInstaller installs exporters on remote nodes
type ExporterInstaller struct {
	executor *ssh.Executor
}

// NewExporterInstaller creates a new exporter installer
func NewExporterInstaller(executor *ssh.Executor) *ExporterInstaller {
	return &ExporterInstaller{executor: executor}
}

// InstallNodeExporter installs node_exporter on a node
func (e *ExporterInstaller) InstallNodeExporter(ctx context.Context, node ssh.Node) error {
	version := Components["node_exporter"].Version

	script := fmt.Sprintf(`
set -e

# Check if already installed
if command -v node_exporter &> /dev/null; then
    echo "node_exporter already installed"
    exit 0
fi

# Download and install
cd /tmp
curl -sLO https://github.com/prometheus/node_exporter/releases/download/v%s/node_exporter-%s.linux-amd64.tar.gz
tar xzf node_exporter-%s.linux-amd64.tar.gz
mv node_exporter-%s.linux-amd64/node_exporter /usr/local/bin/
rm -rf node_exporter-%s.linux-amd64*

# Create systemd service
cat > /etc/systemd/system/node_exporter.service << 'EOF'
[Unit]
Description=Node Exporter
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/node_exporter
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl enable node_exporter
systemctl start node_exporter

echo "node_exporter installed successfully"
`, version, version, version, version, version)

	result := e.executor.Run(ctx, node, script)
	if result.Error != nil {
		return fmt.Errorf("install node_exporter: %w", result.Error)
	}

	return nil
}

// InstallDCGMExporter installs dcgm-exporter on a node
func (e *ExporterInstaller) InstallDCGMExporter(ctx context.Context, node ssh.Node) error {
	script := `
set -e

# Check for NVIDIA GPU
if ! command -v nvidia-smi &> /dev/null; then
    echo "NVIDIA driver not found"
    exit 1
fi

# Check if Docker is available
if command -v docker &> /dev/null; then
    # Use Docker to run dcgm-exporter
    docker pull nvcr.io/nvidia/k8s/dcgm-exporter:3.3.0-3.2.0-ubuntu22.04

    # Stop existing container if any
    docker rm -f dcgm-exporter 2>/dev/null || true

    # Start dcgm-exporter
    docker run -d --name dcgm-exporter \
        --restart always \
        --gpus all \
        -p 9400:9400 \
        nvcr.io/nvidia/k8s/dcgm-exporter:3.3.0-3.2.0-ubuntu22.04

    echo "dcgm-exporter installed via Docker"
else
    # Install DCGM package (requires NVIDIA repository)
    echo "Docker not found. Manual DCGM installation required."
    echo "See: https://developer.nvidia.com/dcgm"
    exit 1
fi
`

	result := e.executor.Run(ctx, node, script)
	if result.Error != nil {
		return fmt.Errorf("install dcgm-exporter: %w", result.Error)
	}

	return nil
}

// CheckExporterStatus checks the status of exporters on a node
func (e *ExporterInstaller) CheckExporterStatus(ctx context.Context, node ssh.Node) (ExporterStatus, error) {
	script := `
echo "node_exporter:"
if systemctl is-active node_exporter &> /dev/null; then
    echo "running"
else
    echo "not_running"
fi

echo "dcgm_exporter:"
if docker ps --format '{{.Names}}' | grep -q dcgm-exporter; then
    echo "running"
elif systemctl is-active dcgm-exporter &> /dev/null; then
    echo "running"
else
    echo "not_running"
fi
`

	result := e.executor.Run(ctx, node, script)
	if result.Error != nil {
		return ExporterStatus{}, result.Error
	}

	status := ExporterStatus{}
	lines := strings.Split(result.Output, "\n")
	for i, line := range lines {
		if strings.Contains(line, "node_exporter:") && i+1 < len(lines) {
			status.NodeExporter = strings.TrimSpace(lines[i+1]) == "running"
		}
		if strings.Contains(line, "dcgm_exporter:") && i+1 < len(lines) {
			status.DCGMExporter = strings.TrimSpace(lines[i+1]) == "running"
		}
	}

	return status, nil
}

// ExporterStatus contains the status of exporters on a node
type ExporterStatus struct {
	NodeExporter bool
	DCGMExporter bool
}
