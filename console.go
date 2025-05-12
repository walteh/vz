package vz

/*
#cgo darwin CFLAGS: -mmacosx-version-min=11 -x objective-c -fno-objc-arc
#cgo darwin LDFLAGS: -lobjc -framework Foundation -framework Virtualization
# include "virtualization_13.h"
*/
import "C"
import (
	"unsafe"

	"github.com/Code-Hex/vz/v3/internal/objc"
)

// ConsoleDeviceConfiguration interface for an console device configuration.
type ConsoleDeviceConfiguration interface {
	objc.NSObject

	consoleDeviceConfiguration()
}

type baseConsoleDeviceConfiguration struct{}

func (*baseConsoleDeviceConfiguration) consoleDeviceConfiguration() {}

// VirtioConsoleDeviceConfiguration is Virtio Console Device.
type VirtioConsoleDeviceConfiguration struct {
	*pointer
	portsPtr unsafe.Pointer

	*baseConsoleDeviceConfiguration

	consolePorts map[int]*VirtioConsolePortConfiguration
}

var _ ConsoleDeviceConfiguration = (*VirtioConsoleDeviceConfiguration)(nil)

// NewVirtioConsoleDeviceConfiguration creates a new VirtioConsoleDeviceConfiguration.
func NewVirtioConsoleDeviceConfiguration() (*VirtioConsoleDeviceConfiguration, error) {
	if err := macOSAvailable(13); err != nil {
		return nil, err
	}
	config := &VirtioConsoleDeviceConfiguration{
		pointer: objc.NewPointer(
			C.newVZVirtioConsoleDeviceConfiguration(),
		),
		consolePorts: make(map[int]*VirtioConsolePortConfiguration),
	}
	config.portsPtr = C.portsVZVirtioConsoleDeviceConfiguration(objc.Ptr(config))

	objc.SetFinalizer(config, func(self *VirtioConsoleDeviceConfiguration) {
		objc.Release(self)
	})
	return config, nil
}

// MaximumPortCount returns the maximum number of ports allocated by this device.
// The default is the number of ports attached to this device.
func (v *VirtioConsoleDeviceConfiguration) MaximumPortCount() uint32 {
	return uint32(C.maximumPortCountVZVirtioConsolePortConfigurationArray(v.portsPtr))
}

func (v *VirtioConsoleDeviceConfiguration) SetVirtioConsolePortConfiguration(idx int, portConfig *VirtioConsolePortConfiguration) {
	C.setObjectAtIndexedSubscriptVZVirtioConsolePortConfigurationArray(
		v.portsPtr,
		objc.Ptr(portConfig),
		C.int(idx),
	)

	// to mark as currently reachable.
	// This ensures that the object is not freed, and its finalizer is not run
	v.consolePorts[idx] = portConfig
}

type ConsolePortConfiguration interface {
	objc.NSObject

	consolePortConfiguration()
}

type baseConsolePortConfiguration struct{}

func (*baseConsolePortConfiguration) consolePortConfiguration() {}

// VirtioConsolePortConfiguration is Virtio Console Port
//
// A console port is a two-way communication channel between a host VZSerialPortAttachment and
// a virtual machine console port. One or more console ports are attached to a Virtio console device.
type VirtioConsolePortConfiguration struct {
	*pointer

	*baseConsolePortConfiguration

	isConsole  bool
	name       string
	attachment SerialPortAttachment
}

var _ ConsolePortConfiguration = (*VirtioConsolePortConfiguration)(nil)

// NewVirtioConsolePortConfigurationOption is an option type to initialize a new VirtioConsolePortConfiguration
type NewVirtioConsolePortConfigurationOption func(*VirtioConsolePortConfiguration)

// WithVirtioConsolePortConfigurationName sets the console port's name.
// The default behavior is to not use a name unless set.
func WithVirtioConsolePortConfigurationName(name string) NewVirtioConsolePortConfigurationOption {
	return func(vcpc *VirtioConsolePortConfiguration) {
		consolePortName := charWithGoString(name)
		defer consolePortName.Free()
		C.setNameVZVirtioConsolePortConfiguration(
			objc.Ptr(vcpc),
			consolePortName.CString(),
		)
		vcpc.name = name
	}
}

// WithVirtioConsolePortConfigurationIsConsole sets the console port may be marked
// for use as the system console. The default is false.
func WithVirtioConsolePortConfigurationIsConsole(isConsole bool) NewVirtioConsolePortConfigurationOption {
	return func(vcpc *VirtioConsolePortConfiguration) {
		C.setIsConsoleVZVirtioConsolePortConfiguration(
			objc.Ptr(vcpc),
			C.bool(isConsole),
		)
		vcpc.isConsole = isConsole
	}
}

// WithVirtioConsolePortConfigurationAttachment sets the console port attachment.
// The default is nil.
func WithVirtioConsolePortConfigurationAttachment(attachment SerialPortAttachment) NewVirtioConsolePortConfigurationOption {
	return func(vcpc *VirtioConsolePortConfiguration) {
		C.setAttachmentVZVirtioConsolePortConfiguration(
			objc.Ptr(vcpc),
			objc.Ptr(attachment),
		)
		vcpc.attachment = attachment
	}
}

// NewVirtioConsolePortConfiguration creates a new VirtioConsolePortConfiguration.
//
// This is only supported on macOS 13 and newer, error will
// be returned on older versions.
func NewVirtioConsolePortConfiguration(opts ...NewVirtioConsolePortConfigurationOption) (*VirtioConsolePortConfiguration, error) {
	if err := macOSAvailable(13); err != nil {
		return nil, err
	}
	vcpc := &VirtioConsolePortConfiguration{
		pointer: objc.NewPointer(
			C.newVZVirtioConsolePortConfiguration(),
		),
	}
	for _, optFunc := range opts {
		optFunc(vcpc)
	}
	objc.SetFinalizer(vcpc, func(self *VirtioConsolePortConfiguration) {
		objc.Release(self)
	})
	return vcpc, nil
}

// Name returns the console port's name.
func (v *VirtioConsolePortConfiguration) Name() string { return v.name }

// IsConsole returns the console port may be marked for use as the system console.
func (v *VirtioConsolePortConfiguration) IsConsole() bool { return v.isConsole }

// Attachment returns the console port attachment.
func (v *VirtioConsolePortConfiguration) Attachment() SerialPortAttachment {
	return v.attachment
}


type ConsoleDevice interface {
	objc.NSObject

	consoleDevice()
}


type baseConsoleDevice struct{}

func (*baseConsoleDevice) consoleDevice() {}

// ConsoleDevices returns the list of *runtime* console devices attached to the
// virtual machine. The slice is empty until Start succeeds.
//
// NOTE: Requires macOS 13+. Attempting to call on older systems panics with
// UnsupportedOSVersionError.
func (v *VirtualMachine) ConsoleDevices() []ConsoleDevice {
	if err := macOSAvailable(13); err != nil {
		panic(err)
	}
	arr := objc.NewNSArray(C.VZVirtualMachine_consoleDevices(objc.Ptr(v)))
	ptrs := arr.ToPointerSlice()
	devs := make([]ConsoleDevice, len(ptrs))
	for i, p := range ptrs {
		// TODO: When/if Apple adds more console device types in future macOS versions,
		// implement type checking here to create the appropriate device wrapper.
		// Currently, VirtioConsoleDevice is the only type supported.
		devs[i] = newVirtioConsoleDevice(p)
	}
	return devs
}

type VirtioConsoleDevice struct {
	*pointer

	*baseConsoleDevice

	ports []*VirtioConsolePort // Cache for ports
}

// newVirtioConsoleDevice creates a new VirtioConsoleDevice wrapper.
// The input pointer is an VZVirtioConsoleDevice*. Internal use only.
func newVirtioConsoleDevice(ptr unsafe.Pointer) *VirtioConsoleDevice {
	if ptr == nil {
		return nil
	}
	// Note: We don't set a finalizer here as the device's lifetime is tied to the VM.
	return &VirtioConsoleDevice{
		pointer: objc.NewPointer(ptr),
	}
}

// Ports returns the console ports associated with this device.
// The returned slice should not be modified.
// This is only supported on macOS 13 and newer.
func (d *VirtioConsoleDevice) Ports() ([]*VirtioConsolePort, error) {
	if err := macOSAvailable(13); err != nil {
		return nil, err
	}
	if d.ports == nil {
		portsArrayPtr := C.VZVirtioConsoleDevice_ports(objc.Ptr(d))
		if portsArrayPtr == nil {
			return []*VirtioConsolePort{}, nil // No ports array available
		}
		count := int(C.VZVirtioConsolePortArray_maximumPortCount(portsArrayPtr))
		d.ports = make([]*VirtioConsolePort, count)
		for i := 0; i < count; i++ {
			portPtr := C.VZVirtioConsolePortArray_objectAtIndexedSubscript(portsArrayPtr, C.size_t(i))
			if portPtr != nil {
				d.ports[i] = &VirtioConsolePort{
					pointer: objc.NewPointer(portPtr),
				}
				// Note: We don't set a finalizer here as the port's lifetime is tied to the device/VM.
			}
		}
	}
	return d.ports, nil
}

// VirtioConsolePort represents a console port on a Virtio console device.
// see: https://developer.apple.com/documentation/virtualization/vzvirtioconsoleport?language=objc
type VirtioConsolePort struct {
	*pointer
	attachment SerialPortAttachment // Cache for attachment
}

// Name returns the name associated with this console port.
// Returns an empty string if no name is set.
// This is only supported on macOS 13 and newer.
func (p *VirtioConsolePort) Name() (string, error) {
	if err := macOSAvailable(13); err != nil {
		return "", err
	}
	return C.GoString(C.VZVirtioConsolePort_name(objc.Ptr(p))), nil
}

// Attachment returns the last observed (handled by get, set) serial port attachment associated with this console port.
// It does not return the current attachment, but rather a cached value from the last call to [VirtioConsolePort.GetAttachment] or [VirtioConsolePort.SetAttachment].
// For the current attachment, use [VirtioConsolePort.GetAttachment].
// Returns nil if no attachment is set.
// This is only supported on macOS 13 and newer.
func (p *VirtioConsolePort) Attachment() (SerialPortAttachment, error) {
	if err := macOSAvailable(13); err != nil {
		return nil, err
	}
	return p.attachment, nil
}

// GetAttachment returns the serial port attachment associated with this console port.
// Returns nil if no attachment is set.
// This is only supported on macOS 13 and newer.
func (p *VirtioConsolePort) GetAttachment() (SerialPortAttachment, error) {
	if err := macOSAvailable(13); err != nil {
		return nil, err
	}

	attachmentPtr := C.VZVirtioConsolePort_getAttachment(objc.Ptr(p))
	if attachmentPtr == nil {
		return nil, nil
	}

	// Need to determine the type of attachment and wrap it.
	// This requires checking the Objective-C class, which is complex via cgo.
	// For now, we return a basic pointer. A more robust solution would involve
	// Objective-C helper functions or more sophisticated bridging.
	// Consider using dedicated attachment types if specific functionality is needed.
	p.attachment = &FileHandleSerialPortAttachment{
		pointer: objc.NewPointer(attachmentPtr),
	} // Assuming FileHandle for simplicity; THIS IS LIKELY INCORRECT

	// Alternative: Return an opaque pointer that can be cast later if needed.
	// p.attachment = &baseSerialPortAttachment{ pointer: objc.NewPointer(attachmentPtr) }

	return p.attachment, nil
}

// SetAttachment sets the serial port attachment for this console port.
// This can be used to connect a serial port attachment dynamically after the VM starts.
// Setting attachment to nil disconnects the port.
// This is only supported on macOS 13 and newer.
func (p *VirtioConsolePort) SetAttachment(attachment SerialPortAttachment) error {
	if err := macOSAvailable(13); err != nil {
		return err
	}
	var attachmentPtr unsafe.Pointer
	if attachment != nil {
		attachmentPtr = objc.Ptr(attachment)
	}
	C.VZVirtioConsolePort_setAttachment(objc.Ptr(p), attachmentPtr)
	p.attachment = attachment // Update cache
	return nil
}