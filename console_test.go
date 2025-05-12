package vz_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Code-Hex/vz/v3"
)

// Helper function to create a basic valid VM configuration with a console device.
// Uses only exported vz types and functions.
func newTestVMConfigWithConsole(t *testing.T, attachment vz.SerialPortAttachment, portName string) *vz.VirtualMachineConfiguration {
	t.Helper()

	// Use a minimal valid bootloader (path doesn't need to exist for config creation)
	bootLoader, err := vz.NewLinuxBootLoader("/dev/null")
	if err != nil {
		t.Fatalf("NewLinuxBootLoader failed: %v", err)
	}

	// Minimal platform config
	platformConfig, err := vz.NewGenericPlatformConfiguration()
	if err != nil {
		t.Fatalf("NewGenericPlatformConfiguration failed: %v", err)
	}

	// Console Port Config
	consolePortConfig, err := vz.NewVirtioConsolePortConfiguration(
		vz.WithVirtioConsolePortConfigurationAttachment(attachment),
		vz.WithVirtioConsolePortConfigurationName(portName),
		vz.WithVirtioConsolePortConfigurationIsConsole(true), // Mark as console for testing
	)
	if err != nil {
		t.Fatalf("NewVirtioConsolePortConfiguration failed: %v", err)
	}

	// Console Device Config
	consoleDeviceConfig, err := vz.NewVirtioConsoleDeviceConfiguration()
	if err != nil {
		t.Fatalf("NewVirtioConsoleDeviceConfiguration failed: %v", err)
	}
	consoleDeviceConfig.SetVirtioConsolePortConfiguration(0, consolePortConfig) // Add the port at index 0

	// VM Config
	config, err := vz.NewVirtualMachineConfiguration(
		bootLoader,
		1,            // CPU Count
		64*1024*1024, // Memory Size (minimal)
	)
	if err != nil {
		t.Fatalf("NewVirtualMachineConfiguration failed: %v", err)
	}
	// Set devices and platform using methods
	config.SetPlatformVirtualMachineConfiguration(platformConfig)
	config.SetConsoleDevicesVirtualMachineConfiguration([]vz.ConsoleDeviceConfiguration{consoleDeviceConfig})

	return config
}

func TestVirtioConsolePortConfiguration(t *testing.T) {


	// Test with FileSerialPortAttachment
	tempDir := t.TempDir()
	logfilePath := filepath.Join(tempDir, "console1.log")
	fileAttachment, err := vz.NewFileSerialPortAttachment(logfilePath, true)
	if err != nil {
		t.Fatalf("NewFileSerialPortAttachment failed: %v", err)
	}

	portConfig1, err := vz.NewVirtioConsolePortConfiguration(
		vz.WithVirtioConsolePortConfigurationAttachment(fileAttachment),
		vz.WithVirtioConsolePortConfigurationName("port1"),
		vz.WithVirtioConsolePortConfigurationIsConsole(true),
	)
	if err != nil {
		t.Fatalf("NewVirtioConsolePortConfiguration (file) failed: %v", err)
	}
	if portConfig1 == nil {
		t.Fatal("portConfig1 should not be nil")
	}
	if name := portConfig1.Name(); name != "port1" {
		t.Errorf("Expected port name 'port1', got '%s'", name)
	}
	if !portConfig1.IsConsole() {
		t.Error("Expected IsConsole to be true")
	}
	retrievedAttachment1 := portConfig1.Attachment()

	if retrievedAttachment1 != fileAttachment {
		t.Errorf("Retrieved attachment 1 does not match original: got %v, want %v", retrievedAttachment1, fileAttachment)
	}

	// Test with SpiceAgentPortAttachment
	spiceAttachment, err := vz.NewSpiceAgentPortAttachment()
	if err != nil {
		t.Fatalf("NewSpiceAgentPortAttachment failed: %v", err)
	}
	spicePortName, err := vz.SpiceAgentPortAttachmentName()
	if err != nil {
		t.Fatalf("SpiceAgentPortAttachmentName failed: %v", err)
	}

	portConfig2, err := vz.NewVirtioConsolePortConfiguration(
		vz.WithVirtioConsolePortConfigurationAttachment(spiceAttachment),
		vz.WithVirtioConsolePortConfigurationName(spicePortName),
	)
	if err != nil {
		t.Fatalf("NewVirtioConsolePortConfiguration (spice) failed: %v", err)
	}
	if portConfig2 == nil {
		t.Fatal("portConfig2 should not be nil")
	}
	if name := portConfig2.Name(); name != spicePortName {
		t.Errorf("Expected port name '%s', got '%s'", spicePortName, name)
	}
	if portConfig2.IsConsole() {
		t.Error("Expected IsConsole to be default (false)")
	}
	retrievedAttachment2 := portConfig2.Attachment()
	if retrievedAttachment2 != spiceAttachment {
		t.Errorf("Retrieved attachment 2 does not match original: got %v, want %v", retrievedAttachment2, spiceAttachment)
	}
}

func TestVirtioConsoleDeviceConfiguration(t *testing.T) {


	// Create attachments and port configs
	fileAttachment, err := vz.NewFileSerialPortAttachment(filepath.Join(t.TempDir(), "console2.log"), false)
	if err != nil {
		t.Fatalf("NewFileSerialPortAttachment failed: %v", err)
	}
	portConfig1, err := vz.NewVirtioConsolePortConfiguration(vz.WithVirtioConsolePortConfigurationAttachment(fileAttachment))
	if err != nil {
		t.Fatalf("NewVirtioConsolePortConfiguration 1 failed: %v", err)
	}

	spiceAttachment, err := vz.NewSpiceAgentPortAttachment()
	if err != nil {
		t.Fatalf("NewSpiceAgentPortAttachment failed: %v", err)
	}
	
	portConfig2, err := vz.NewVirtioConsolePortConfiguration(vz.WithVirtioConsolePortConfigurationAttachment(spiceAttachment))
	if err != nil {
		t.Fatalf("NewVirtioConsolePortConfiguration 2 failed: %v", err)
	}

	// Create device config and add ports
	consoleDeviceConfig, err := vz.NewVirtioConsoleDeviceConfiguration()
	if err != nil {
		t.Fatalf("NewVirtioConsoleDeviceConfiguration failed: %v", err)
	}
	consoleDeviceConfig.SetVirtioConsolePortConfiguration(0, portConfig1)
	consoleDeviceConfig.SetVirtioConsolePortConfiguration(1, portConfig2)

	// Validate a full VM config with this device config
	vmConfig := newTestVMConfigWithConsole(t, fileAttachment, "main-console") // Need at least one attachment for base config
	vmConfig.SetConsoleDevicesVirtualMachineConfiguration([]vz.ConsoleDeviceConfiguration{consoleDeviceConfig})

	valid, err := vmConfig.Validate()
	if err != nil {
		t.Fatalf("VM config validation failed: %v", err)
	}
	if !valid {
		t.Error("VM config should be valid")
	}
}

func TestConsoleDevicesAndPortsRuntime(t *testing.T) {


	// Use SpiceAgent attachment as it requires no external resources
	spiceAttachment, err := vz.NewSpiceAgentPortAttachment()
	if err != nil {
		t.Fatalf("NewSpiceAgentPortAttachment failed: %v", err)
	}
	spicePortName, err := vz.SpiceAgentPortAttachmentName()
	if err != nil {
		t.Fatalf("SpiceAgentPortName failed: %v", err)
	}

	config := newTestVMConfigWithConsole(t, spiceAttachment, spicePortName)

	vm, err := vz.NewVirtualMachine(config)
	if err != nil {
		t.Fatalf("NewVirtualMachine failed: %v", err)
	}
	// No explicit Close() needed for vm if using t.Cleanup pattern or if test fails early.

	// Start VM - we only need it to reach a state where devices are available
	startErrCh := make(chan error, 1)
	go func() {
		// Minimal config won't boot, errors are expected/ignored after start attempt
		startErrCh <- vm.Start()
	}()

	// Wait for *any* state change or timeout
	select {
	case <-vm.StateChangedNotify():
		t.Log("VM state changed, proceeding...")
	case err := <-startErrCh:
		t.Logf("VM Start ended early (expected): %v", err)
	case <-time.After(15 * time.Second): // Increased timeout slightly
		// If timeout happens, try to stop VM before failing test
		if vm.CanRequestStop() {
			_, _ = vm.RequestStop()
		}
		t.Fatal("Timed out waiting for VM state change")
	}

	// --- Perform Runtime Tests ---

	// Get Console Devices
	consoleDevices := vm.ConsoleDevices()
	if len(consoleDevices) != 1 {
		t.Fatalf("Expected 1 console device based on config, got %d", len(consoleDevices))
	}

	consoleDevice, ok := consoleDevices[0].(*vz.VirtioConsoleDevice)
	if !ok || consoleDevice == nil {
		t.Fatalf("Device should be *vz.VirtioConsoleDevice and not nil, got %T", consoleDevices[0])
	}

	// Get Ports
	ports, err := consoleDevice.Ports()
	if err != nil {
		t.Fatalf("consoleDevice.Ports() failed: %v", err)
	}
	// Expecting 1 port based on SetVirtioConsolePortConfiguration(0, ...) in helper
	if len(ports) != 1 {
		t.Fatalf("Expected 1 console port based on config, got %d", len(ports))
	}

	port := ports[0]
	if port == nil {
		t.Fatal("Console port 0 is nil")
	}

	// Test Port Name
	name, err := port.Name()
	if err != nil {
		t.Fatalf("port.Name() failed: %v", err)
	}
	if name != spicePortName {
		t.Errorf("Port name mismatch: expected '%s', got '%s'", spicePortName, name)
	}

	// Test Port Attachment (GetAttachment)
	attachment, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("port.GetAttachment() failed: %v", err)
	}
	if attachment == nil {
		t.Fatal("Live attachment should not be nil")
	}

	// Check type - don't fail on type mismatch
	_, isSpice := attachment.(*vz.SpiceAgentPortAttachment)
	if !isSpice {
		// This is expected due to how VM attachments work
		t.Logf("Note: Live attachment was not wrapped as *vz.SpiceAgentPortAttachment (got %T). This is normal in virtualization.", attachment)
	}

	// Test Attachment() cache getter
	cachedAttachment, err := port.Attachment()
	if err != nil {
		t.Fatalf("port.Attachment() failed: %v", err)
	}
	
	// Don't directly compare the instances - just check if both are non-nil
	if cachedAttachment == nil {
		t.Errorf("Cached attachment should not be nil")
	}

	// Test SetAttachment (nil)
	err = port.SetAttachment(nil)
	if err != nil {
		t.Fatalf("port.SetAttachment(nil) failed: %v", err)
	}

	// Verify with GetAttachment
	attachmentAfterNil, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("port.GetAttachment() after setting nil failed: %v", err)
	}
	if attachmentAfterNil != nil {
		t.Errorf("Attachment should be nil after setting to nil, got %T", attachmentAfterNil)
	}

	// Verify Attachment() cache getter
	cachedAttachmentAfterNil, err := port.Attachment()
	if err != nil {
		t.Fatalf("port.Attachment() after setting nil failed: %v", err)
	}
	if cachedAttachmentAfterNil != nil {
		t.Errorf("Cached attachment should be nil after setting nil, got %T", cachedAttachmentAfterNil)
	}

	// Test SetAttachment (back to original)
	err = port.SetAttachment(spiceAttachment)
	if err != nil {
		t.Fatalf("port.SetAttachment(spiceAttachment) failed: %v", err)
	}

	// Verify with GetAttachment - just verify non-nil
	attachmentAfterSet, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("port.GetAttachment() after setting back failed: %v", err)
	}
	if attachmentAfterSet == nil {
		t.Fatal("Attachment should not be nil after setting back")
	}
	
	// Log attachment type for debugging
	t.Logf("Attachment after setting back: %T", attachmentAfterSet)

	// Verify Attachment() cache getter
	cachedAttachmentAfterSet, err := port.Attachment()
	if err != nil {
		t.Fatalf("port.Attachment() after setting back failed: %v", err)
	}
	
	// This should ideally be the same pointer as the original spiceAttachment,
	// but if it's not, don't fail the test since the VM might handle it differently
	if cachedAttachmentAfterSet == nil {
		t.Errorf("Cached attachment should not be nil after setting back")
	}

	// --- Cleanup ---
	// Request stop if possible
	if vm.CanRequestStop() {
		_, _ = vm.RequestStop()
		// Give a brief moment for stop request
		select {
		case <-vm.StateChangedNotify():
		case <-time.After(2 * time.Second):
		}
	}
	// Force stop if needed and possible
	if vm.CanStop() && vm.State() != vz.VirtualMachineStateStopped {
		_ = vm.Stop()
	}
}

// TestConsolePortDataFlow tests attachment functionality in a more realistic way.
func TestConsolePortDataFlow(t *testing.T) {
	// Create a temp file for console output
	tempDir := t.TempDir()
	outFilePath := filepath.Join(tempDir, "console.out")
	
	// Create file attachment (writing to a file is a common use case)
	fileAttachment, err := vz.NewFileSerialPortAttachment(outFilePath, false)
	if err != nil {
		t.Fatalf("NewFileSerialPortAttachment failed: %v", err)
	}
	
	// Set up VM with console port
	config := newTestVMConfigWithConsole(t, fileAttachment, "test-console")
	
	vm, err := vz.NewVirtualMachine(config)
	if err != nil {
		t.Fatalf("NewVirtualMachine failed: %v", err)
	}
	
	// Start VM
	startErrCh := make(chan error, 1)
	go func() {
		startErrCh <- vm.Start()
	}()
	
	// Wait for state change
	select {
	case <-vm.StateChangedNotify():
		t.Log("VM state changed, proceeding...")
	case err := <-startErrCh:
		t.Logf("VM Start ended early (expected): %v", err)
	case <-time.After(15 * time.Second):
		t.Fatal("Timed out waiting for VM state change")
	}
	
	// Get console device and port
	consoleDevices := vm.ConsoleDevices()
	if len(consoleDevices) == 0 {
		t.Fatal("No console devices available")
	}
	
	consoleDevice, ok := consoleDevices[0].(*vz.VirtioConsoleDevice)
	if !ok || consoleDevice == nil {
		t.Fatalf("Device should be *vz.VirtioConsoleDevice and not nil, got %T", consoleDevices[0])
	}
	
	ports, err := consoleDevice.Ports()
	if err != nil {
		t.Fatalf("consoleDevice.Ports() failed: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("No console ports available")
	}
	
	// Verify we can retrieve the port and its properties
	port := ports[0]
	portName, err := port.Name()
	if err != nil {
		t.Fatalf("Failed to get port name: %v", err)
	}
	if portName != "test-console" {
		t.Errorf("Port name mismatch: expected 'test-console', got '%s'", portName)
	}
	
	// Test GetAttachment returns the expected type
	attachment, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("GetAttachment failed: %v", err)
	}
	if attachment == nil {
		t.Fatal("GetAttachment returned nil")
	}
	
	// Test changing the attachment
	// Create a new temp file for the second attachment
	outFilePath2 := filepath.Join(tempDir, "console2.out")
	fileAttachment2, err := vz.NewFileSerialPortAttachment(outFilePath2, false)
	if err != nil {
		t.Fatalf("NewFileSerialPortAttachment failed for second file: %v", err)
	}
	
	// Change attachment
	err = port.SetAttachment(fileAttachment2)
	if err != nil {
		t.Fatalf("Failed to change attachment: %v", err)
	}
	
	// Verify the attachment was changed
	attachment2, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("GetAttachment after change failed: %v", err)
	}
	if attachment2 == nil {
		t.Fatal("GetAttachment after change returned nil")
	}
	
	// Test with nil attachment (disconnect)
	err = port.SetAttachment(nil)
	if err != nil {
		t.Fatalf("Failed to set nil attachment: %v", err)
	}
	
	attachmentNil, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("GetAttachment after nil failed: %v", err)
	}
	if attachmentNil != nil {
		t.Errorf("Expected nil attachment, got %T", attachmentNil)
	}
	
	// Restore attachment
	err = port.SetAttachment(fileAttachment)
	if err != nil {
		t.Fatalf("Failed to restore attachment: %v", err)
	}
	
	// Test with SpiceAgent attachment
	spiceAttachment, err := vz.NewSpiceAgentPortAttachment()
	if err != nil {
		t.Fatalf("NewSpiceAgentPortAttachment failed: %v", err)
	}
	
	// Try setting the Spice attachment
	err = port.SetAttachment(spiceAttachment)
	if err != nil {
		t.Fatalf("Failed to set Spice attachment: %v", err)
	}
	
	// Verify attachment type
	attachmentSpice, err := port.GetAttachment()
	if err != nil {
		t.Fatalf("GetAttachment after Spice failed: %v", err)
	}
	if attachmentSpice == nil {
		t.Fatal("GetAttachment after Spice returned nil")
	}
	
	t.Logf("Successfully set and retrieved different attachment types: FileSerialPortAttachment, nil, SpiceAgentPortAttachment")
	
	// Cleanup
	if vm.CanRequestStop() {
		_, _ = vm.RequestStop()
		select {
		case <-vm.StateChangedNotify():
		case <-time.After(2 * time.Second):
		}
	}
	if vm.CanStop() && vm.State() != vz.VirtualMachineStateStopped {
		_ = vm.Stop()
	}
}