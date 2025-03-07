package dcom

type AuthenticationLevel uint32

const (
	// Default COM authentication level
	AuthenticationLevelDefault AuthenticationLevel = 0
	// No authentication
	AuthenticationLevelNone AuthenticationLevel = 1
	// Connect level authentication
	AuthenticationLevelConnect AuthenticationLevel = 2
	// Call level authentication
	AuthenticationLevelCall AuthenticationLevel = 3
	// Packet level authentication
	AuthenticationLevelPacket AuthenticationLevel = 4
	// Packet integrity level authentication
	AuthenticationLevelPacketIntegrity AuthenticationLevel = 5
	// Packet privacy level authentication
	AuthenticationLevelPacketPrivacy AuthenticationLevel = 6
)

func (a AuthenticationLevel) String() string {
	switch a {
	case AuthenticationLevelDefault:
		return "Default"
	case AuthenticationLevelNone:
		return "None"
	case AuthenticationLevelConnect:
		return "Connect"
	case AuthenticationLevelCall:
		return "Call"
	case AuthenticationLevelPacket:
		return "Packet"
	case AuthenticationLevelPacketIntegrity:
		return "PacketIntegrity"
	case AuthenticationLevelPacketPrivacy:
		return "PacketPrivacy"
	default:
		return "Unknown"
	}
}

func (a AuthenticationLevel) IsValid() bool {
	return a >= AuthenticationLevelDefault && a <= AuthenticationLevelPacketPrivacy
}
