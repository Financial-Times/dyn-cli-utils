package dyn

import "github.com/nesv/go-dynect/dynect"

type dynectService struct {
	client *dynect.Client
}

func NewDynectService(client *dynect.Client) dynectService {
	return dynectService{
		client: client,
	}
}
