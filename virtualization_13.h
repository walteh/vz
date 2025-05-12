//
//  virtualization_13.h
//
//  Created by codehex.
//

#pragma once

#import "virtualization_helper.h"
#import <Virtualization/Virtualization.h>

/* macOS 13 API */
void setConsoleDevicesVZVirtualMachineConfiguration(void *config, void *consoleDevices);

void *newVZEFIBootLoader();
void setVariableStoreVZEFIBootLoader(void *bootLoaderPtr, void *variableStore);
void *newVZEFIVariableStorePath(const char *variableStorePath);
void *newCreatingVZEFIVariableStoreAtPath(const char *variableStorePath, void **error);

void *newVZGenericMachineIdentifierWithBytes(void *machineIdentifierBytes, int len);
nbyteslice getVZGenericMachineIdentifierDataRepresentation(void *machineIdentifierPtr);
void *newVZGenericMachineIdentifier();
void setMachineIdentifierVZGenericPlatformConfiguration(void *config, void *machineIdentifier);

void *newVZUSBMassStorageDeviceConfiguration(void *attachment);
void *newVZVirtioGraphicsDeviceConfiguration();
void setScanoutsVZVirtioGraphicsDeviceConfiguration(void *graphicsConfiguration, void *scanouts);
void *newVZVirtioGraphicsScanoutConfiguration(NSInteger widthInPixels, NSInteger heightInPixels);

void *newVZVirtioConsoleDeviceConfiguration();
void *portsVZVirtioConsoleDeviceConfiguration(void *consoleDevice);
uint32_t maximumPortCountVZVirtioConsolePortConfigurationArray(void *ports);
void *getObjectAtIndexedSubscriptVZVirtioConsolePortConfigurationArray(void *portsPtr, int portIndex);
void setObjectAtIndexedSubscriptVZVirtioConsolePortConfigurationArray(void *portsPtr, void *portConfig, int portIndex);

void *newVZVirtioConsolePortConfiguration();
void setNameVZVirtioConsolePortConfiguration(void *consolePortConfig, const char *name);
void setIsConsoleVZVirtioConsolePortConfiguration(void *consolePortConfig, bool isConsole);
void setAttachmentVZVirtioConsolePortConfiguration(void *consolePortConfig, void *serialPortAttachment);
void *newVZSpiceAgentPortAttachment();
void setSharesClipboardVZSpiceAgentPortAttachment(void *attachment, bool sharesClipboard);
const char *getSpiceAgentPortName();

void startWithOptionsCompletionHandler(void *machine, void *queue, void *options, uintptr_t cgoHandle);

const char *getMacOSGuestAutomountTag();

void setMaximumTransmissionUnitVZFileHandleNetworkDeviceAttachment(void *attachment, NSInteger mtu);


/* VZVirtioConsoleDevice */
void *VZVirtualMachine_consoleDevices(void *machine);

void *VZVirtioConsoleDevice_ports(void *consoleDevice);
void VZVirtioConsoleDevice_setDelegate(void *consoleDevice,
                                       uintptr_t cgoHandle,
                                       VZVirtioConsolePortCallback didOpen,
                                       VZVirtioConsolePortCallback didClose);

 typedef void (*VZVirtioConsolePortCallback)(uintptr_t cgoHandle,
                                            void *consoleDevice /* VZVirtioConsoleDevice* */,
                                            void *port          /* VZVirtioConsolePort* */);

@interface VZVirtioConsoleDeviceDelegateImpl : NSObject <VZVirtioConsoleDeviceDelegate>
- (instancetype)initWithHandle:(uintptr_t)cgoHandle
            didOpenCallback:(VZVirtioConsolePortCallback)didOpen
           didCloseCallback:(VZVirtioConsolePortCallback)didClose;
@end

/* VZVirtioConsolePortArray */
size_t VZVirtioConsolePortArray_maximumPortCount(void *portArray);
void *VZVirtioConsolePortArray_objectAtIndexedSubscript(void *portArray, size_t index);

/* VZVirtioConsolePort */
const char *VZVirtioConsolePort_name(void *port);
void *VZVirtioConsolePort_setAttachment(void *port, void *attachment);    
void *VZVirtioConsolePort_getAttachment(void *port);

