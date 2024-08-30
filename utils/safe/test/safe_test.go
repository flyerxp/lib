package test

import (
	"fmt"
	"github.com/flyerxp/lib/v2/utils/safe"
	jwt "github.com/golang-jwt/jwt/v5"
	"testing"
)

func TestJwt(t *testing.T) {
	key := []byte("aaaaaaaaaaa")
	j := safe.CreateJwt(key, func(opts *safe.JwtOptions) {
	})
	token, err := j.Encode(jwt.MapClaims{
		"temp":  22.5,
		"speed": 25.2,
	})
	fmt.Println(token, err)
	//token = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzcGVlZCI6MjUuMSwidGVtcCI6MjIuNX0.f4ynIicEuDZO7TbtQ5-JA0jRrwcbJzrnHN33fno4nsLZAVL0wv25gZJ2EF8zUkx_VHm0xRUpt4sQhAMQIndOSg"
	temp, e := j.Decode(token)
	fmt.Println(temp, e)
}
func TestAes(t *testing.T) {
	k := []byte("abcdefghigklmnop")
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	a := safe.CreateAesL(k, iv)
	out, _ := a.Encrypt([]byte("abcdefgabcdefgaasdfasdfasdf"))
	fmt.Println(out)
	fmt.Println(a.Decrypt(out))
}
func TestAes2(t *testing.T) {
	k := []byte("tWqPVPszz5SOLEzI")
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	a := safe.CreateAesL(k, iv)
	//out, _ := a.Encrypt([]byte("abcdefgabcdefgaasdfasdfasdf"))
	out := "tXoGv7G0PT3QMD2Z11zACmEleXPRDili2yRH01VUCPLMR4OhLfrogHVhHYoEtOKibGWjZEyIrP8UophYsF+Q5jEcqU3tQ8iz2v62TzVO1FYkQwd9cL/jxW4CgTgQg6BSHeuFMqNc8AhxP87Vg0npFA+fHr9LiVr/Qxeqc7/Rc6aSc2AL8Qp6gQJwZWvRFtYkQl5uMWvUkJwo/d55Qx2/u5l5G63BIiz4yj/1Jnx/6cA="
	fmt.Println(out)
	fmt.Println(a.Decrypt(out))
}
