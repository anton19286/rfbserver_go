//Server handshake handler functions

package rfb

import (
	"io"
)

func serverProtocolVersionHandler(c io.ReadWriter) (string, error) {
	n, err := c.Write([]byte(v8))
	if err != nil {
		return "", err
	}
	b := make([]byte, len(v8))
	n, err = c.Read(b)
	return string(b[:n]), err
}

func serverSecurityTypeHandler(c io.ReadWriter) (byte, error) {
	_, err := c.Write([]byte{1, 1})
	if err != nil {
		return 0, err
	}
	b := [1]byte{}
	_, err = c.Read(b[:])
	return b[0], err
}

func serverSecurityResultHandler(c io.ReadWriter) error {
	_, err := c.Write([]byte{0, 0, 0, 0})
	return err
}

func serverInitHandler(c io.ReadWriter, sc ServerConfig) (bool, error) {
	b := [1]byte{}
	_, err := c.Read(b[:])
	if err != nil {
		return false, err
	}
	shareFlag := false
	if b[0] > 0 {
		shareFlag = true
	}
	cpf := sc.PixelFormat
	pf := []byte{
		cpf.BPP,
		cpf.Depth,
		cpf.BigEndian,
		cpf.TrueColour,
		byte(cpf.RedMax >> 8),
		byte(cpf.RedMax),
		byte(cpf.GreenMax >> 8),
		byte(cpf.GreenMax),
		byte(cpf.BlueMax >> 8),
		byte(cpf.BlueMax),
		cpf.RedShift,
		cpf.GreenShift,
		cpf.BlueShift,
		0, 0, 0, //padding
	}
	data := []byte{byte(sc.Width >> 8), byte(sc.Width),
		byte(sc.Height >> 8), byte(sc.Height)}
	data = append(data, pf...)
	n := []byte(sc.DesktopName)
	l := len(n)
	data = append(data, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
	data = append(data, n...)
	_, err = c.Write(data)
	return shareFlag, err
}
