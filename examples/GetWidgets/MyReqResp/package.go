package MyReqResp

import (
	//"github.com/bketelsen/skynet/skylib" // SkynetRequest/Response should be here.
	"os"
)


type SkynetRequest struct {
	Params map[string]interface{}
	Body   []byte
}

func (self *SkynetRequest) BodyAsString() (body string, err os.Error) {
	return string(self.Body), nil
}

type SkynetResponse struct {
	Errors []string
	Body   []byte
}

func (self *SkynetResponse) WriteString(s string) {
	self.Body = []byte(s)
}


type NameRepo struct {
	first, last, middle string
}

func (self *NameRepo) find(index string) string {
	println(index)
	switch index {
	case "0":
		return self.first
	case "1":
		return self.middle
	case "2":
		return self.last
	}
	return ""
}

var male_repo = &NameRepo{first: "paulo", middle: "moura", last: "suzart"}
var female_repo = &NameRepo{first: "carolina", middle: "giraldo", last: "valle"}


type UnisexService struct {

}

func (*UnisexService) HandleMale(req SkynetRequest, resp *SkynetResponse) (err os.Error) {
	i, _ := req.BodyAsString()
	s := male_repo.find(i)
	resp.WriteString(s)
	return
}

func (*UnisexService) HandleFemale(req SkynetRequest, resp *SkynetResponse) (err os.Error) {
	i, _ := req.BodyAsString()
	s := female_repo.find(i)
	resp.WriteString(s)
	return
}
