package webcam

import (
	"bytes"
	"encoding/binary"
	"github.com/fabian-z/webcam/ioctl"
	"golang.org/x/sys/unix"
	"unsafe"
)

type v4l2_format struct {
	_type uint32
	union [200]uint8
}

type v4l2_buffer struct {
	index     uint32
	_type     uint32
	bytesused uint32
	flags     uint32
	field     uint32
	timestamp unix.Timeval
	timecode  v4l2_timecode
	sequence  uint32
	memory    uint32
	union     [4]uint8
	length    uint32
	reserved2 uint32
	reserved  uint32
}

func setImageFormat(fd uintptr, formatcode *uint32, width *uint32, height *uint32) (err error) {

	format := &v4l2_format{
		_type: V4L2_BUF_TYPE_VIDEO_CAPTURE,
	}

	pix := v4l2_pix_format{
		Width:       *width,
		Height:      *height,
		Pixelformat: *formatcode,
		Field:       V4L2_FIELD_ANY,
	}

	pixbytes := &bytes.Buffer{}
	err = binary.Write(pixbytes, binary.LittleEndian, pix)

	if err != nil {
		return
	}

	copy(format.union[:], pixbytes.Bytes())

	err = ioctl.Ioctl(fd, VIDIOC_S_FMT, uintptr(unsafe.Pointer(format)))

	if err != nil {
		return
	}

	pixReverse := &v4l2_pix_format{}
	err = binary.Read(bytes.NewBuffer(format.union[:]), binary.LittleEndian, pixReverse)

	if err != nil {
		return
	}

	*width = pixReverse.Width
	*height = pixReverse.Height
	*formatcode = pixReverse.Pixelformat

	return

}

func waitForFrame(fd uintptr, timeout uint32) (count int, err error) {

	for {

		fds := &FdSet{}
		fds.Set(fd)

		tv := &unix.Timeval{}
		tv.Sec = int32(timeout)

		countReturn, _, errno := unix.Syscall6(unix.SYS_SELECT, uintptr(fd+1), uintptr(unsafe.Pointer(fds)), uintptr(0), uintptr(0), uintptr(unsafe.Pointer(tv)), 0)

		count = int(countReturn)

		if errno != 0 {
			err = errno
		}

		if count < 0 {

			if err == unix.EINTR {
				continue
			}

			return

		}

		return

	}

}
