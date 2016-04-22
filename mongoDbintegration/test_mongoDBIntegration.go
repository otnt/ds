package main

import (
	"fmt"
	"github.com/otnt/ds/mongoDBintegration"
	"labix.org/v2/mgo/bson"
)

func main() {
	mongoSession := mongoDBintegration.EstablishSession()

	/* Insert fake data */
	objectID := mongoDBintegration.InsertPicture(mongoSession, "https://goo.gl/lc37Hh", "Prathi", "prathi", "nil")
	fmt.Println("Inserted the picture \n")
	objectID_string := objectID.Hex()
	fmt.Println("The object id is %s", objectID_string)

}
