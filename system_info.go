package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GetServerInfo returns the server hostname and IP address
func GetServerInfo() (string, string) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Error fetching hostname: %v", err)
		hostname = "Unknown Host"
	}

	ips, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Error fetching IP address: %v", err)
		return hostname, "Unknown IP"
	}

	for _, addr := range ips {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return hostname, ipnet.IP.String()
		}
	}
	return hostname, "Unknown IP"
}

// GetDiskUsage returns disk usage percentage
func GetDiskUsage() (int, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_LogicalDisk | Select-Object Size,FreeSpace,DeviceID")
	} else {
		cmd = exec.Command("df", "-h", "/")
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute command: %w", err)
	}

	if runtime.GOOS == "windows" {
		return parseWindowsDiskUsage(string(output))
	}

	return parseUnixDiskUsage(string(output))
}

// GetCPUUsage returns CPU usage percentage
func GetCPUUsage() (float64, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_Processor | Measure-Object -Property LoadPercentage -Average | Select-Object Average")
	} else {
		cmd = exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute command: %w", err)
	}

	if runtime.GOOS == "windows" {
		return parseWindowsCPUUsage(string(output))
	}

	return parseUnixCPUUsage(string(output))
}

// GetMemoryUsage returns memory usage percentage
func GetMemoryUsage() (float64, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_OperatingSystem | Select-Object TotalVisibleMemorySize,FreePhysicalMemory")
	} else {
		cmd = exec.Command("sh", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0}'")
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute command: %w", err)
	}

	if runtime.GOOS == "windows" {
		return parseWindowsMemoryUsage(string(output))
	}

	return parseUnixMemoryUsage(string(output))
}

// parseUnixDiskUsage parses Unix df output
func parseUnixDiskUsage(output string) (int, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	usage := fields[4]
	usagePercent, err := strconv.Atoi(strings.TrimSuffix(usage, "%"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse disk usage percentage: %w", err)
	}

	return usagePercent, nil
}

// parseUnixCPUUsage parses Unix CPU usage output
func parseUnixCPUUsage(output string) (float64, error) {
	cpuUsage, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CPU usage: %w", err)
	}
	return cpuUsage, nil
}

// parseUnixMemoryUsage parses Unix memory usage output
func parseUnixMemoryUsage(output string) (float64, error) {
	memoryUsage, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory usage: %w", err)
	}
	return memoryUsage, nil
}

// parseWindowsDiskUsage parses Windows disk usage output
func parseWindowsDiskUsage(output string) (int, error) {
	// TODO: Implement proper Windows disk usage parsing
	// This is a simplified implementation
	return 0, fmt.Errorf("windows disk usage parsing not implemented")
}

// parseWindowsCPUUsage parses Windows CPU usage output
func parseWindowsCPUUsage(output string) (float64, error) {
	// TODO: Implement proper Windows CPU usage parsing
	// This is a simplified implementation
	return 0.0, fmt.Errorf("windows CPU usage parsing not implemented")
}

// parseWindowsMemoryUsage parses Windows memory usage output
func parseWindowsMemoryUsage(output string) (float64, error) {
	// TODO: Implement proper Windows memory usage parsing
	// This is a simplified implementation
	return 0.0, fmt.Errorf("windows memory usage parsing not implemented")
}
