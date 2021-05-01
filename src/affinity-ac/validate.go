package main

import (
	"fmt"
	"net/http"
	"os"
)

func (vac *myServerHandler) valserve(w http.ResponseWriter, r *http.Request) {
	
	fmt.Fprintf(os.Stdout, "starting validation request")
}