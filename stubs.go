package libauth

// OwnerImpl ???
type OwnerImpl struct {
	Sig *Signer
}

// NewOwner ???
func NewOwner(signer Signer) *OwnerImpl {
	o := new(OwnerImpl)

	o.Sig = &signer

	return o
}

// Signer ???
func (o *OwnerImpl) Signer() *Signer {
	return o.Sig
}

// ClientImpl ???
type ClientImpl struct {
	Ver    *Verifier
	Connec *Con
	Dig    *Digest
}

// NewClient ???
func NewClient(verifier Verifier, con Con) *ClientImpl {
	c := new(ClientImpl)

	c.Ver = &verifier
	c.Connec = &con

	return c
}

// Verifier ???
func (c *ClientImpl) Verifier() *Verifier {
	return c.Ver
}

// Con ???
func (c *ClientImpl) Con() *Con {
	return c.Connec
}

// Digest ???
func (c *ClientImpl) Digest() *Digest {
	return c.Dig
}

// Update ???
func (c *ClientImpl) Update(digest *Digest) {
	c.Dig = digest
}

// Query ???
func (c *ClientImpl) Query(query *Query) (bool, *Resp) {
	res := (*c.Con()).Query(query)

	v := (*c.Verifier()).Verify(res.Data.(*VO))

	if !v {return v, nil}

	return v, res
}

// ServerImpl ???
type ServerImpl struct {
	Auther *Authenticator
	Dig    *Digest
	Dat    *Data
}

// NewServer ???
func NewServer(authenticator Authenticator) *ServerImpl {
	s := new(ServerImpl)

	s.Auther = &authenticator

	return s
}

// Update ???
func (s *ServerImpl) Update(digest *Digest, data *Data) {
	s.Dig = digest
	s.Dat = data
}

// Query ???
func (s *ServerImpl) Query(query *Query) *Resp {
	panic("todo")
}

// Authenticator ???
func (s *ServerImpl) Authenticator() *Authenticator {
	return s.Auther
}

// Digest ???
func (s *ServerImpl) Digest() *Digest {
	return s.Dig
}

// Data ???
func (s *ServerImpl) Data() *Data {
	return s.Dat
}

// VerifierStub ???
type VerifierStub struct {
	Vo *VO
}

// NewVerifierStub ???
func NewVerifierStub(vo *VO) *VerifierStub {
	v := new(VerifierStub)

	v.Vo = vo

	return v
}

// Verify ???
func (v *VerifierStub) Verify(vo *VO) bool {
	return vo.Data == v.Vo.Data
}

// AuthenticatorStub ???
type AuthenticatorStub struct {
	Vo *VO
}

// NewAuthStub ???
func NewAuthStub(vo *VO) *AuthenticatorStub {
	a := new(AuthenticatorStub)

	a.Vo = vo

	return a
}

// Auth ???
func (a *AuthenticatorStub) Auth(resp *Resp) *Resp {
	r := new(Resp)
	r.Type = stub
	r.Data = a.Vo

	return r
}

// SignerStub ???
type SignerStub struct {
	Sig *Data
}

// NewSignerStub ???
func NewSignerStub(sig *Data) *SignerStub {
	s := new(SignerStub)

	s.Sig = sig

	return s
}

// Sign ???
func (s *SignerStub) Sign(data *Data) *Data {
	return s.Sig
}

// ConStub ???
type ConStub struct {
	Serv *Server
}

// NewConStub ???
func NewConStub(server Server) *ConStub {
	c := new(ConStub)

	c.Serv = &server

	return c
}

// Query ???
func (c *ConStub) Query(query *Query) *Resp {
	return (*c.Serv).Query(query)
}
