package chainutils

// ScriptType takes in a script and returns "P2PKH", "P2WPKH", "P2SH", "P2WSH", "P2PK", "INVALID" denoting the type of transaction it is, and the relevant non opcode data
func ScriptType(pkScript []byte) (string, []byte) {

	if len(pkScript) == 22 && pkScript[0] == 0x00 && pkScript[1] == 0x14 {
		return "P2WPKH", pkScript[2:22]
	}

	if len(pkScript) == 23 && pkScript[0] == 0xa9 && pkScript[1] == 0x14 && pkScript[22] == 0x87 {
		return "P2SH", pkScript[2:22]
	}

	if len(pkScript) == 25 && pkScript[0] == 0x76 && pkScript[1] == 0xa9 && pkScript[2] == 0x14 && pkScript[23] == 0x88 && pkScript[24] == 0xac {
		return "P2PKH", pkScript[3:23]
	}

	if len(pkScript) == 34 && pkScript[0] == 0x00 && pkScript[1] == 0x20 {
		return "P2WSH", pkScript[2:34]
	}

	if len(pkScript) == 67 && pkScript[0] == 0x41 && pkScript[66] == 0xac {
		return "P2PK", pkScript[1:66]
	}

	return "INVALID", nil
}
