package libauth

// Owner ???
type Owner interface {
	Signer() *Signer
}

// Client ???
type Client interface{
	Verifier() *Verifier	// Verifier ???
	Con() *Con				// Con ???
	Digest() *Digest		// Digest ???
	Update(digest *Digest)	// Update ???
}

// Server ???
type Server interface{
	Update(digest *Digest, data *Data) // Update ???
	Query(query *Query) *Resp // Query ???
	Authenticator() *Authenticator // Authenticator ???
	Digest() *Digest // Digest ???
	Data() *Data // Data ???
}

// Verifier ???
type Verifier interface {
	Verify(vo *VO) bool // Verify ???
}

// Authenticator ???
type Authenticator interface {
	Auth(resp *Resp) *Resp // Auth ???
}

// Signer ???
type Signer interface {
	Sign(data *Data) *Resp // Sign ???
}

// Con ???
type Con interface{
	Query(query *Query) // Handle ???
}

// QType ???
type QType int

const ()

// Query ???
type Query interface {
	Type() QType      // Type ???
	Data() interface{} // Data ???
}

// VOType ???
type VOType int

const ()

// VO ???
type VO interface {
	Type() VOType      // Type ???
	Data() interface{} // Data ???
}

// RespType ???
type RespType int

const ()

// Resp ???
type Resp interface {
	Type() RespType      // Type ???
	Data() interface{} // Data ???
}

// DigestType ???
type DigestType int

const ()

// Digest ???
type Digest interface {
	Type() DigestType      // Type ???
	Data() interface{} // Data ???
}

// DataType ???
type DataType int

const ()

// Data ???
type Data interface {
	Type() DataType      // Type ???
	Data() interface{} // Data ???
}