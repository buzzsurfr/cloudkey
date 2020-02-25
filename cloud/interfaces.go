package cloud

// Profile interface provides methods to implement to work with profiles from the cloud types
type Profile interface {
	RotateKey()
}
