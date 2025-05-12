package vz

/*
#cgo darwin CFLAGS: -mmacosx-version-min=11 -x objective-c -fno-objc-arc
#cgo darwin LDFLAGS: -lobjc -framework Foundation -framework Virtualization
# include "virtualization_13.h"
*/
import "C"
import (
	"github.com/Code-Hex/vz/v3/internal/objc"
)

// SpiceAgentPortAttachment is an attachment point that enables
// the Spice clipboard sharing capability.
//
// see: https://developer.apple.com/documentation/virtualization/vzspiceagentportattachment?language=objc
type SpiceAgentPortAttachment struct {
	*pointer

	*baseSerialPortAttachment
}

var _ SerialPortAttachment = (*SpiceAgentPortAttachment)(nil)

// NewSpiceAgentPortAttachment creates a new Spice agent port attachment.
//
// This is only supported on macOS 13 and newer, error will
// be returned on older versions.
func NewSpiceAgentPortAttachment() (*SpiceAgentPortAttachment, error) {
	if err := macOSAvailable(13); err != nil {
		return nil, err
	}
	spiceAgent := &SpiceAgentPortAttachment{
		pointer: objc.NewPointer(
			C.newVZSpiceAgentPortAttachment(),
		),
	}
	objc.SetFinalizer(spiceAgent, func(self *SpiceAgentPortAttachment) {
		objc.Release(self)
	})
	return spiceAgent, nil
}

// SetSharesClipboard sets enable the Spice agent clipboard sharing capability.
func (s *SpiceAgentPortAttachment) SetSharesClipboard(enable bool) {
	C.setSharesClipboardVZSpiceAgentPortAttachment(
		objc.Ptr(s),
		C.bool(enable),
	)
}

// SharesClipboard returns whether clipboard sharing is enabled for the Spice agent.
func (a *SpiceAgentPortAttachment) SharesClipboard() (bool, error) {
	if err := macOSAvailable(13); err != nil {
		return false, err
	}
	return bool(C.getSharesClipboardVZSpiceAgentPortAttachment(objc.Ptr(a))), nil
}

// SpiceAgentPortAttachmentName returns the Spice agent port name.
// It is always the same value.
func SpiceAgentPortAttachmentName() (string, error) {
	if err := macOSAvailable(13); err != nil {
		return "", err
	}
	cstring := (*char)(C.getSpiceAgentPortName())
	return cstring.String(), nil
}
