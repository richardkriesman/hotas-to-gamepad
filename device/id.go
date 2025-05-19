package device

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"github.com/holoplot/go-evdev"
)

type PersistentID string

func createPersistentID(device *evdev.InputDevice) PersistentID {
	var identifiers []byte

	// add input ID info to the pool
	inputID, err := device.InputID()
	if err == nil {
		for _, value := range []uint16{inputID.BusType, inputID.Vendor, inputID.Product, inputID.Version} {
			binary.LittleEndian.AppendUint16(identifiers, value)
		}
	}

	// add name if available
	name, err := device.Name()
	if err == nil {
		identifiers = append(identifiers, []byte(name)...)
	}

	// add unique id if available
	uniqueID, err := device.UniqueID()
	if err == nil {
		identifiers = append(identifiers, []byte(uniqueID)...)
	}

	// TODO: add more identifiers? still produces duplicates with Razer Naga V2 Hyperspeed mouse

	// create hash from the identifier pool
	hash := sha256.Sum256(identifiers)
	return PersistentID(hex.EncodeToString(hash[:]))
}
