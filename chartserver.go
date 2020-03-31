/*
   Copyright 2020 Jacobo Tarrío Barreiro (http://jacobo.tarrio.org)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"flag"
	"hamchart/chartgen"
	"log"
	"net/http"

	"github.com/markbates/pkger"
)

var serverAddress = flag.String("server_address", "127.0.0.1:8080", "Address of the HTTP server.")

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(pkger.Dir("/web")))
	mux.HandleFunc("/chart", chartgen.ChartHandler)

	server := &http.Server{
		Addr:    *serverAddress,
		Handler: mux,
	}
	log.Printf("Server is now ready to listen on %s", *serverAddress)
	log.Fatal(server.ListenAndServe())
}
