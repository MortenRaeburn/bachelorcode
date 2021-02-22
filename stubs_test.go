package libauth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testStartLabel(t *testing.T) {
	fmt.Println("---" + t.Name() + "---")
	fmt.Println()
}

func testEndLabel() {
	fmt.Println()
	fmt.Println()
}

func TestPositive(t *testing.T) {
	testStartLabel(t)
	defer testEndLabel()

	assert := assert.New(t)

	vo := new(VO)
	vo.Type = stub
	vo.Data = 69

	voVer := new(VO)
	voVer.Type = stub
	voVer.Data = 69

	auther := NewAuthStub(vo)

	serv := NewServer(auther)

	con := NewConStub(serv)

	ver := NewVerifierStub(voVer)

	client := NewClient(ver, con)

	dig := new(Digest)
	dig.Type = stub
	dig.Data = 9999

	data := new(Data)
	data.Type = stub
	data.Data = 123

	client.Update(dig)
	serv.Update(dig, data)

	v, res := client.Query(nil)

	assert.Equal(true, v, "Expected response to be positive")
	assert.Equal(data.Data, res.Data, "Expected response data to equal initial data")

}
