package igd

import "github.com/NebulousLabs/go-upnp"

var gIGD *upnp.IGD
var gErr error

func GetIGD() (igd *upnp.IGD, err error) {
	// Don't retry discovery
	if gErr != nil {
		err = gErr
		return
	}

	if gIGD != nil {
		igd = gIGD
		return
	}

	gIGD, err = upnp.Discover()
	if err != nil {
		gErr = err
		return
	}
	igd = gIGD
	return
}
