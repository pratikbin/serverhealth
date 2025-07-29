package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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

// GetDiskUsage returns disk usage percentage using native Go
func GetDiskUsage() (int, error) {
	if runtime.GOOS == "windows" {
		return getWindowsDiskUsage()
	}
	return getUnixDiskUsage()
}

// GetCPUUsage returns CPU usage percentage using native Go
func GetCPUUsage() (float64, error) {
	if runtime.GOOS == "windows" {
		return getWindowsCPUUsage()
	}
	return getUnixCPUUsage()
}

// GetMemoryUsage returns memory usage percentage using native Go
func GetMemoryUsage() (float64, error) {
	if runtime.GOOS == "windows" {
		return getWindowsMemoryUsage()
	}
	return getUnixMemoryUsage()
}

// getUnixDiskUsage reads /proc/mounts and /proc/stat for disk usage
func getUnixDiskUsage() (int, error) {
	// Read /proc/mounts to find root filesystem
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return 0, fmt.Errorf("failed to open /proc/mounts: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var rootDevice string
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == "/" {
			rootDevice = fields[0]
			break
		}
	}

	if rootDevice == "" {
		return 0, fmt.Errorf("could not find root filesystem")
	}

	// Get filesystem statistics using statfs
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate usage percentage
	totalBlocks := stat.Blocks
	freeBlocks := stat.Bfree
	usedBlocks := totalBlocks - freeBlocks
	usagePercent := int((float64(usedBlocks) / float64(totalBlocks)) * 100)

	return usagePercent, nil
}

// getUnixCPUUsage reads /proc/stat for CPU usage
func getUnixCPUUsage() (float64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, fmt.Errorf("failed to open /proc/stat: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var cpuLine string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			cpuLine = line
			break
		}
	}

	if cpuLine == "" {
		return 0, fmt.Errorf("could not find CPU line in /proc/stat")
	}

	// Parse CPU line: cpu  user nice system idle iowait irq softirq steal guest guest_nice
	fields := strings.Fields(cpuLine)
	if len(fields) < 5 {
		return 0, fmt.Errorf("invalid CPU line format")
	}

	// Calculate total and idle time
	var total, idle uint64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			continue
		}
		total += val
		if i == 4 { // idle time is the 5th field (index 4)
			idle = val
		}
	}

	if total == 0 {
		return 0, fmt.Errorf("invalid CPU statistics")
	}

	// Calculate CPU usage percentage
	usagePercent := 100.0 - (float64(idle) / float64(total) * 100.0)
	return usagePercent, nil
}

// getUnixMemoryUsage reads /proc/meminfo for memory usage
func getUnixMemoryUsage() (float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var total, available uint64
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					total = val
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					available = val
				}
			}
		}
	}

	if total == 0 {
		return 0, fmt.Errorf("could not read memory information")
	}

	// Calculate memory usage percentage
	used := total - available
	usagePercent := (float64(used) / float64(total)) * 100.0
	return usagePercent, nil
}

// Windows implementations using WMI via PowerShell
func getWindowsDiskUsage() (int, error) {
	// This is a simplified implementation using PowerShell
	// In a production environment, you might want to use a proper WMI library
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject -Class Win32_LogicalDisk | Where-Object {$_.DeviceID -eq 'C:'} | Select-Object @{Name='Usage';Expression={[math]::Round((($_.Size - $_.FreeSpace) / $_.Size) * 100)}}")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get Windows disk usage: %w", err)
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "Usage" && !strings.HasPrefix(line, "---") {
			usage, err := strconv.Atoi(line)
			if err == nil {
				return usage, nil
			}
		}
	}

	return 0, fmt.Errorf("could not parse Windows disk usage")
}

func getWindowsCPUUsage() (float64, error) {
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject -Class Win32_Processor | Measure-Object -Property LoadPercentage -Average | Select-Object -ExpandProperty Average")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get Windows CPU usage: %w", err)
	}

	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Windows CPU usage: %w", err)
	}

	return usage, nil
}

func getWindowsMemoryUsage() (float64, error) {
	cmd := exec.Command("powershell", "-Command",
		"$os = Get-WmiObject -Class Win32_OperatingSystem; [math]::Round((($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) / $os.TotalVisibleMemorySize) * 100, 2)")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get Windows memory usage: %w", err)
	}

	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Windows memory usage: %w", err)
	}

	return usage, nil
}
